package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/rivo/tview"
	"github.com/sqweek/dialog"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var globalFrom *tview.Form

const FireButtonName = "Fire!"
const CeaseButtonName = "Cease"
const UseTestFun = false

type StatisticsData struct {
	SedMsg7211Count uint64
	SedMsg7221Count uint64
}

func handleMsg(ctx context.Context, id int, setting *Setting, staticsData *StatisticsData) {
	client, err := NewCFMQClient(setting.Server, setting.Username, setting.Password)
	if err != nil {
		AppLogger.Printf("Worker %d create CFMQ clinet error: %s\n", id, err)
	}
	client.SedQueueTips = setting.SedQueueTips
	err = client.CreateQueue(client.SedQueueTips)
	if err != nil {
		AppLogger.Printf("Worker %d create sed queue tips error: %s\n", id, err)
	}
	defer client.Logout()
	go client.HeartBeat(ctx)
	// 从setting.FilePath目录地址获取目录下的所有xml文件，并循环处理xml文件，发送报文
	AppLogger.Printf("读取文件：%s\n", setting.FilePath)
	filepath.Walk(setting.FilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查是否为 XML 文件
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".xml" {
			// 获取单个 XML 文件
			msg, err := processSingleXMLFile(path)
			if err != nil {
				return err
			}
			// 获取报文头
			msgNo, _ := getXMLFieldValue(msg, "MsgNo")
			AppLogger.Printf("获取报文类型：%s\n", msgNo)
			if err != nil {
				return err
			}
			if msgNo == "7221" {
				// 替换MsgID
				msg = replaceXMLField(msg, "MsgID", GenerateUniqueTipsId)
				// 替换MsgRef
				msg = replaceXMLField(msg, "MsgRef", GenerateUniqueTipsId)
				// 替换DES
				msg = replaceXMLFieldWithValue(msg, "DES", "333333333333")
				// 替换PackNo
				msg = replaceXMLField(msg, "PackNo", GeneratePackNo)

				// 获取国库代码
				treCode, _ := getXMLFieldValue(msg, "DrawBackTreCode")

				// 发送7221报文给mq
				client.SendTipsMsg(msg, treCode)

				atomic.AddUint64(&staticsData.SedMsg7221Count, 1)
			}
			if msgNo == "7211" {
				now := time.Now()
				// 替换MsgID
				msg = replaceXMLField(msg, "MsgID", GenerateUniqueTipsId)
				// 替换MsgRef
				msg = replaceXMLField(msg, "MsgRef", GenerateUniqueTipsId)
				// 替换DES
				msg = replaceXMLFieldWithValue(msg, "DES", "333333333333")
				// 替换PackNo
				msg = replaceXMLField(msg, "PackNo", GeneratePackNo)
				// 替换EntrustDate
				msg = replaceXMLFieldWithValue(msg, "EntrustDate", now.Format("20060102"))
				// 替换TraNo
				msg = replaceXMLField(msg, "TraNo", GenerateTraNo)
				// 替换TaxVouNo
				msg = replaceXMLField(msg, "TaxVouNo", GenerateTaxVouNo)
				// 替换BillDate
				msg = replaceXMLFieldWithValue(msg, "BillDate", now.Format("20060102"))

				// 获取国库代码
				treCode, _ := getXMLFieldValue(msg, "PayeeTreCode")

				// 发送7211报文
				client.SendTipsMsg(msg, treCode)

				atomic.AddUint64(&staticsData.SedMsg7211Count, 1)
			}
		}
		return nil
	})
	AppLogger.Printf("Worker %d finished!\n", id)
}

func processSingleXMLFile(filePath string) (string, error) {
	// 读取 XML 文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		AppLogger.Printf("Error reading XML file %s: %v", filePath, err)
		return "", err
	}

	return string(data), err
}

func processSingleXMLFileToStruct(filePath string) (*BaseMsg, error) {
	// 读取 XML 文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		AppLogger.Printf("Error reading XML file %s: %v", filePath, err)
		return nil, err
	}

	return parseMsgBody(string(data))
}

// ValueGenerator 定义生成值的函数类型
type ValueGenerator func() string

/**
* 替换XML文件中的字段值，传入数值生成函数，返回替换后的XML文件内容
 */
func replaceXMLField(xmlData string, fieldName string, generator ValueGenerator) string {
	startTag := "<" + fieldName + ">"
	endTag := "</" + fieldName + ">"

	result := xmlData
	offset := 0

	for {
		startIdx := strings.Index(result[offset:], startTag)
		if startIdx == -1 {
			break
		}
		startIdx += offset

		endIdx := strings.Index(result[startIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// 生成新的值
		newValue := generator()

		// 替换字段值
		prefix := result[:startIdx+len(startTag)]
		suffix := result[endIdx:]
		result = prefix + newValue + suffix

		// 更新偏移量以继续查找下一个匹配项
		offset = startIdx + len(startTag) + len(newValue) + len(endTag)

		// 防止无限循环的安全检查
		if offset >= len(result) {
			break
		}
	}

	return result
}

// replaceXMLFieldWithValue 使用固定值替换（保持向后兼容）
func replaceXMLFieldWithValue(xmlData, fieldName, newValue string) string {
	return replaceXMLField(xmlData, fieldName, func() string { return newValue })
}

// 从xml文件中读取某个fieldName对应的值
func getXMLFieldValue(xmlData, fieldName string) (string, error) {
	startTag := "<" + fieldName + ">"
	endTag := "</" + fieldName + ">"
	startIdx := strings.Index(xmlData, startTag)
	if startIdx == -1 {
		return "", errors.New("field not found")
	}
	endIdx := strings.Index(xmlData, endTag)
	if endIdx == -1 {
		return "", errors.New("end tag not found")
	}
	fieldValue := xmlData[startIdx+len(startTag) : endIdx]
	return fieldValue, nil
}

func parseMsgHeader(msgStr string) (*MsgHeader, error) {
	headerStartIndex := strings.Index(msgStr, "{")
	if headerStartIndex < 0 {
		return nil, errors.New("can't find the msg start symbol {")
	}
	headerStr := msgStr[headerStartIndex : headerStartIndex+158]
	revMsgHeader := &MsgHeader{}
	revMsgHeader.ParseHeader(headerStr)
	return revMsgHeader, nil
}

func parseMsgTipsHead(msgStr string) (*MsgHead, error) {
	msgHead := &MsgHead{}
	contentStartIndex := strings.Index(msgStr, "<?xml")
	if contentStartIndex < 0 {
		return nil, errors.New("can't find the xml start symbol <?xml")
	}
	contentStr := msgStr[contentStartIndex:]

	err := xml.Unmarshal([]byte(contentStr), &msgHead)
	return msgHead, err
}

func parseMsgBody(msgStr string) (*BaseMsg, error) {
	baseMsg := &BaseMsg{}
	contentStartIndex := strings.Index(msgStr, "<?xml")
	if contentStartIndex < 0 {
		return nil, errors.New("can't find the xml start symbol <?xml")
	}
	contentStr := msgStr[contentStartIndex:]

	err := xml.Unmarshal([]byte(contentStr), &baseMsg)
	return baseMsg, err

}

func sendMsg990(revMsgHeader *MsgHeader, revMsgBody *BaseMsg, client *CFMQClient) {
	now := time.Now()
	msgId := GenerateUniqueId()
	msg990header := MsgHeader{
		OrigSender:   revMsgHeader.OrigReceiver,
		OrigReceiver: revMsgHeader.OrigSender,
		OrigSendTime: now.Format("20060102150405"),
		MsgType:      "ctbs.990.001.01",
		MsgId:        msgId,
		OrgnlMsgId:   revMsgHeader.MsgId,
	}
	msg990 := CTBS990Msg{
		MsgId:      msgId,
		OrgnlSndr:  strings.TrimSpace(revMsgHeader.OrigSender),
		OrgnlSndDt: revMsgHeader.OrigSendTime,
		OrgnlMsgId: revMsgHeader.MsgId,
		OrgnlMT:    revMsgHeader.MsgType,
		RtnCd:      "CT010000",
	}

	client.SendMsg(msg990header.BuildHeader()+msg990.Build990Msg(), revMsgBody.InstgPty)
}

func sendMsg900(revMsgHeader *MsgHeader, revMsgBody *BaseMsg, client *CFMQClient) {
	now := time.Now()
	msgId := GenerateUniqueId()
	header := &MsgHeader{
		OrigSender:   revMsgHeader.OrigReceiver,
		OrigReceiver: revMsgHeader.OrigSender,
		OrigSendTime: now.Format("20060102150405"),
		MsgType:      "ctbs.900.001.01",
		MsgId:        msgId,
		OrgnlMsgId:   revMsgHeader.MsgId,
	}
	res := &CTBS900Msg{
		MsgId:         msgId,
		CreDtTm:       now.Format("2006-01-02T15:04:05"),
		InstgPty:      revMsgBody.InstdPty,
		InstdPty:      revMsgBody.InstgPty,
		OrgnlMsgId:    revMsgBody.MsgId,
		OrgnlInstgPty: revMsgBody.InstgPty,
		OrgnlMT:       revMsgBody.MsgType,
		PrcSts:        "PR00",
	}
	client.SendMsg(header.BuildHeader()+res.Build900Msg(), revMsgBody.InstgPty)
}

func buildMsgId() string {
	//msgId := time.Now().Format("20060102150405")
	timestamp := time.Now().UnixMilli()
	msgId := strconv.FormatInt(timestamp, 10)
	randId := strconv.Itoa(rand.Int() % 100000)

	for range 5 - len(randId) {
		msgId += "0"
	}
	msgId += randId
	return msgId
}

func displayStatistics(ctx context.Context, data *StatisticsData, list *tview.TextView, app *tview.Application) {
	data.SedMsg7211Count = 0
	data.SedMsg7221Count = 0
	for {
		select {
		case <-ctx.Done():
			app.QueueUpdateDraw(func() {
				list.Clear()
				fmt.Fprintf(list, "Sed tips.7211 [%d]\n", data.SedMsg7211Count)
				fmt.Fprintf(list, "Sed tips.7221 [%d]\n", data.SedMsg7221Count)

				fmt.Fprintf(list, "\nCurrent Time is %s\n", time.Now().Format("2006-01-02T15:04:05"))
				fmt.Fprintf(list, "执行结束...")
			})
			AppLogger.Printf("Display worker stopping.\n")
			return
		default:
			app.QueueUpdateDraw(func() {
				list.Clear()
				fmt.Fprintf(list, "Sed tips.7211 [%d]\n", data.SedMsg7211Count)
				fmt.Fprintf(list, "Sed tips.7221 [%d]\n", data.SedMsg7221Count)

				fmt.Fprintf(list, "\nCurrent Time is %s\n", time.Now().Format("2006-01-02T15:04:05"))
				time.Sleep(500 * time.Microsecond)
			})
		}
	}
}

// 构建目录树的辅助函数
func buildTree(path string, parent *tview.TreeNode) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// 先添加目录，再添加文件
	dirs := []os.DirEntry{}
	filesList := []os.DirEntry{}

	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			filesList = append(filesList, file)
		}
	}

	// 添加目录节点
	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir.Name())
		node := tview.NewTreeNode(dir.Name())
		node.SetReference(dirPath)
		node.SetSelectable(true)
		// 为目录添加占位符子节点，表示可以展开
		if hasSubDirs(dirPath) {
			placeholder := tview.NewTreeNode("")
			node.AddChild(placeholder)
		}
		parent.AddChild(node)
	}
}

// 检查目录是否包含子目录
func hasSubDirs(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			return true
		}
	}
	return false
}

func showDirectoryBrowser(app *tview.Application, mainFlex *tview.Flex, setting *Setting) {
	// 创建文件浏览器组件
	fileBrowser := tview.NewTreeView()
	fileBrowser.SetBorder(true).SetTitle("Select Directory")

	// 获取当前目录并设置根节点
	currentDir, _ := os.Getwd()
	root := tview.NewTreeNode(currentDir)
	root.SetReference(currentDir)
	fileBrowser.SetRoot(root)
	fileBrowser.SetCurrentNode(root)

	// 加载目录内容
	loadDirectoryContent(currentDir, root)

	// 设置节点选择事件
	fileBrowser.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}

		path := reference.(string)
		fileInfo, err := os.Stat(path)
		if err != nil {
			return
		}

		if fileInfo.IsDir() {
			// 展开或收缩目录
			if len(node.GetChildren()) == 0 {
				loadDirectoryContent(path, node)
			} else {
				node.SetExpanded(!node.IsExpanded())
			}
		}
	})

	// 创建操作按钮
	buttons := tview.NewFlex()
	okButton := tview.NewButton("OK")
	cancelButton := tview.NewButton("Cancel")

	// 确认按钮事件 - 选择目录并返回主界面
	okButton.SetSelectedFunc(func() {
		currentNode := fileBrowser.GetCurrentNode()
		if currentNode != nil {
			reference := currentNode.GetReference()
			if reference != nil {
				path := reference.(string)
				fileInfo, err := os.Stat(path)
				if err == nil && fileInfo.IsDir() {
					app.QueueUpdateDraw(func() {
						// 更新 FilePath 输入框
						filePathField := globalFrom.GetFormItemByLabel("FilePath").(*tview.InputField)
						filePathField.SetText(path)
						setting.FilePath = path
						// 返回主界面
						app.SetRoot(mainFlex, true)
					})
				}
			}
		}
	})

	// 取消按钮事件 - 直接返回主界面
	cancelButton.SetSelectedFunc(func() {
		app.QueueUpdateDraw(func() {
			app.SetRoot(mainFlex, true)
		})
	})

	// 布局按钮
	buttons.AddItem(okButton, 0, 1, true)
	buttons.AddItem(cancelButton, 0, 1, false)

	// 整体布局
	browserLayout := tview.NewFlex()
	browserLayout.SetDirection(tview.FlexRow)
	browserLayout.AddItem(fileBrowser, 0, 1, true)
	browserLayout.AddItem(buttons, 3, 0, false)

	// 显示文件浏览器
	app.QueueUpdateDraw(func() {
		app.SetRoot(browserLayout, true)
	})
}

func loadDirectoryContent(path string, parent *tview.TreeNode) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// 只加载目录，过滤文件
	for _, file := range files {
		if file.IsDir() {
			fullPath := filepath.Join(path, file.Name())
			node := tview.NewTreeNode(file.Name())
			node.SetReference(fullPath)
			node.SetSelectable(true)
			// 添加占位符子节点表示可以展开
			placeholder := tview.NewTreeNode("")
			node.AddChild(placeholder)
			parent.AddChild(node)
		}
	}
}

func main() {
	var setting = &Setting{}
	setting.Load()
	setting.IsRunning = false

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	// 初始化
	ctx, cancel = context.WithCancel(context.Background())
	statisticdata := &StatisticsData{}
	var testClient *CFMQClient

	app := tview.NewApplication()

	statisticsList := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)
	statisticsList.SetBorder(true).SetTitle("Msg Statistics")

	form := tview.NewForm().
		AddInputField("ServerUrl", setting.Server, 50, nil, func(text string) { setting.Server = text }).
		AddInputField("Username", setting.Username, 50, nil, func(text string) { setting.Username = text }).
		AddPasswordField("Password", setting.Password, 50, '*', func(text string) { setting.Password = text }).
		AddInputField("SedQueueTips", setting.SedQueueTips, 50, nil, func(text string) { setting.SedQueueTips = text }).
		AddInputField("FilePath", setting.FilePath, 50, nil, func(text string) { setting.FilePath = text }).
		AddButton("Browse...", func() {
			// 调用系统文件选择对话框
			go func() {
				dir, err := dialog.Directory().Title("Choose XML Directory").Browse()
				if err == nil {
					app.QueueUpdateDraw(func() {
						filePathField := globalFrom.GetFormItemByLabel("FilePath").(*tview.InputField)
						filePathField.SetText(dir)
						setting.FilePath = dir
					})
				}
			}()
		}).
		//AddButton("Browse...", func() {
		//	showDirectoryBrowser(app, nil, setting)
		//}).
		AddButton(FireButtonName, func() {
			if globalFrom == nil {
				return
			}
			// 立即更新按钮状态为"运行中"
			button := globalFrom.GetButton(globalFrom.GetButtonIndex(FireButtonName))
			button.SetLabel("Running...")
			button.SetDisabled(true) // 按钮会置灰并禁用
			// 创建新的 context
			ctx, cancel = context.WithCancel(context.Background())
			go displayStatistics(ctx, statisticdata, statisticsList, app)
			handleMsg(ctx, 1, setting, statisticdata)
			cancel()
			button.SetLabel(FireButtonName)
			button.SetDisabled(false) // 恢复正常状态
		}).
		AddButton("Quit", func() {
			setting.Save()
			if testClient != nil {
				testClient.Logout()
			}
			cancel()
			app.Stop()
		})
	form.SetBorder(true).SetTitle("Settings").SetTitleAlign(tview.AlignCenter)
	globalFrom = form

	flex := tview.NewFlex().
		AddItem(statisticsList, 0, 1, false).
		AddItem(form, 0, 1, true)

	// 然后修改 Browse 按钮 - 需要重新获取 form 并修改按钮
	//form.GetButton(form.GetButtonIndex("Browse...")).SetSelectedFunc(func() {
	//	showDirectoryBrowser(app, flex, setting)
	//})

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
