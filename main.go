package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/rivo/tview"
	"github.com/sqweek/dialog"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var globalFrom *tview.Form
var convertForm *tview.Form
var decryptForm *tview.Form

const FireButtonName = "执行"
const CeaseButtonName = "Cease"
const Decrypt = "解密"
const Convert = "转换"

const UseTestFun = false

type StatisticsData struct {
	SedMsg7211Count  uint64
	SedMsg7221Count  uint64
	DecryptFileCount uint64
	ConvertFileCount uint64
	Detail7211Count  uint64
	Detail7221Count  uint64
}

/**
* 解密文件
 */
func decryptFiles(originalFilePath string, encKey string, staticsData *StatisticsData) (string, error) {
	// 获取解密后目录：与setting.OriginalFilePathh平级的新目录
	AppLogger.Printf("创建解密后的文件目录")
	baseDir := filepath.Dir(originalFilePath)                                                                              // 获取setting.OriginalFilePath的上级目录
	targetDir := filepath.Join(baseDir, filepath.Base(originalFilePath)+"_decrypted_"+time.Now().Format("20060102150405")) // 创建平级的decrypted目录

	// 确保目标目录存在
	err := os.RemoveAll(targetDir)
	if err != nil {
		AppLogger.Printf("清空目录失败: %v", err)
		return "", err
	}
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		AppLogger.Printf("创建目录失败: %v", err)
		return "", err
	}
	err = filepath.Walk(originalFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			AppLogger.Printf("遍历原始数据目录失败, 原始数据目录：%s, 错误原因：%v", originalFilePath, err)
			return err
		}
		AppLogger.Printf("开始解密文件: %s", path)
		// 如果为xml文件，直接将此文件复制到setting.OriginalFilePath目录平级的目录里面
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".xml" {
			// 构建目标文件路径
			targetFilePath := filepath.Join(targetDir, filepath.Base(path))

			// 复制文件
			err = copyFile(path, targetFilePath)
			if err != nil {
				AppLogger.Printf("复制文件失败 %s 到 %s: %v", path, targetFilePath, err)
				return err
			} else {
				atomic.AddUint64(&staticsData.DecryptFileCount, 1)
			}

			AppLogger.Printf("文件已复制: %s -> %s", path, targetFilePath)
		}
		// 如果为enc文件，需要解密后转为xml文件，再保存到setting.OriginalFilePath目录平级的目录里面
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".enc" {
			decryptedText, err := DecryptFile(path, encKey)
			if err != nil {
				AppLogger.Printf("解密文件失败: %v", err)
				return err
			}
			// 构建目标文件路径，将.enc扩展名改为.xml
			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".xml"
			targetFilePath := filepath.Join(targetDir, baseName)
			// 保存解密后的文件
			err = SaveFile(decryptedText, targetFilePath)
			if err != nil {
				AppLogger.Printf("保存解密文件失败 %s: %v", targetFilePath, err)
				return err
			} else {
				atomic.AddUint64(&staticsData.DecryptFileCount, 1)
			}

			AppLogger.Printf("文件已解密并保存: %s -> %s", path, targetFilePath)
		}
		return nil
	})
	if err != nil {
		AppLogger.Printf("解密文件失败: %v", err)
		return "", err
	}
	return targetDir, nil
}

// copyFile 实现文件复制功能
func copyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制内容
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func convertFiles(decryptedFilePath string, originalFilePath string, payeeOpBkCode string, staticsData *StatisticsData) (string, error) {
	// 获取解密后目录：与settingDecryptedFilePath平级的新目录
	AppLogger.Printf("创建转换后的文件目录")
	baseDir := filepath.Dir(decryptedFilePath) // 获取上级目录
	// 创建平级的convert目录, 文件名+时间（年月日时分秒）
	targetDir := filepath.Join(baseDir, filepath.Base(originalFilePath)+"_convert_"+time.Now().Format("20060102150405"))

	// 如果目标目录存在则清空目录内容，如果不存在则创建
	err := os.RemoveAll(targetDir)
	if err != nil {
		AppLogger.Printf("清空目录失败: %v", err)
		return "", err
	}
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		AppLogger.Printf("创建目录失败: %v", err)
		return "", err
	}

	err = filepath.Walk(decryptedFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			AppLogger.Printf("遍历解密数据目录失败, 解密数据目录：%s, 错误原因：%v", decryptedFilePath, err)
			return err
		}
		AppLogger.Printf("开始转换文件: %s", path)
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".xml" {
			outputDatas, err := ConvertMsg(path, payeeOpBkCode)
			if err != nil {
				AppLogger.Printf("转换文件失败: %v", err)
				return err
			}
			count := 0
			for _, outputData := range outputDatas {
				AppLogger.Printf("开始保存转换后的文件")
				// 构建目标文件路径
				targetFilePath := filepath.Join(targetDir, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))+fmt.Sprintf("_%d.xml", count))
				if len(outputDatas) == 1 {
					targetFilePath = filepath.Join(targetDir, filepath.Base(path))
				}
				// 保存文件
				err := ioutil.WriteFile(targetFilePath, []byte(outputData.Item8), 0644)
				if err != nil {
					AppLogger.Printf("写入文件失败: %v", err)
				} else {
					atomic.AddUint64(&staticsData.ConvertFileCount, 1)
					// 将明细数进行累加
					atomic.AddUint64(&staticsData.Detail7211Count, uint64(outputData.Detail7211Count))
					atomic.AddUint64(&staticsData.Detail7221Count, uint64(outputData.Detail7221Count))
				}
				count++
			}
		}
		//else {
		//	AppLogger.Printf("转换文件不是xml文件: %s", path)
		//}

		return nil
	})
	if err != nil {
		AppLogger.Printf("转换文件失败: %v", err)
		return "", err
	}
	return targetDir, nil
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

	// 解密
	decryptedFilePath, err := decryptFiles(setting.FilePath, setting.EncKey, staticsData)
	if decryptedFilePath == "" || err != nil {
		AppLogger.Printf("Worker %d decrypt error: %s\n", id, err)
		return
	}

	// 等待3秒
	time.Sleep(3 * time.Second)

	// 转换
	convertedFilePath, err := convertFiles(decryptedFilePath, setting.FilePath, setting.PayeeOpBkCode, staticsData)
	if convertedFilePath == "" || err != nil {
		AppLogger.Printf("Worker %d convert error: %s\n", id, err)
		return
	}

	// 等待3秒
	time.Sleep(3 * time.Second)

	//AppLogger.Printf("创建转换后的文件目录")
	//// 获取上级目录
	//baseDir := filepath.Dir(setting.FilePath)
	//// 创建平级的convert目录, 文件名+时间（年月日时分秒）
	//targetDir := filepath.Join(baseDir, filepath.Base(setting.FilePath)+"_convert_"+time.Now().Format("20060102150405"))
	//
	//// 如果目标目录存在则清空目录内容，如果不存在则创建
	//err := os.RemoveAll(targetDir)
	//if err != nil {
	//	AppLogger.Printf("清空目录失败: %v", err)
	//}
	//err = os.MkdirAll(targetDir, 0755)
	//if err != nil {
	//	AppLogger.Printf("创建目录失败: %v", err)
	//}
	//
	//AppLogger.Printf("读取文件：%s\n", setting.FilePath)
	//filepath.Walk(setting.FilePath, func(path string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//	// 如果为enc文件，则解密；如果为xml文件，则直接获取
	//	decryptedText := ""
	//	if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".enc" {
	//		text, err := DecryptFile(path, setting.EncKey)
	//		if err != nil {
	//			AppLogger.Printf("解密文件失败: %v", err)
	//			return err
	//		}
	//		// 转为GBK格式
	//		gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	//		decryptedText, err = gbkEncoder.String(text)
	//		if err != nil {
	//			return fmt.Errorf("GBK编码失败: %v", err)
	//		}
	//	}
	//	if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".xml" {
	//		// 获取单个 XML 文件
	//		decryptedText, err = processSingleXMLFile(path)
	//		if err != nil {
	//			return err
	//		}
	//	}
	err = filepath.Walk(convertedFilePath, func(path string, info os.FileInfo, err error) error {
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
			// 等待100ms
			time.Sleep(100 * time.Millisecond)
			if msgNo == "7221" {
				// 【替换操作提前到转换的时候了，这里不需要再替换了】
				//// 替换MsgID
				//msg = replaceXMLField(msg, "MsgID", GenerateUniqueTipsId)
				//// 替换MsgRef
				//msg = replaceXMLField(msg, "MsgRef", GenerateUniqueTipsId)
				//// 替换DES
				//msg = replaceXMLFieldWithValue(msg, "DES", "333333333333")
				//// 替换PackNo
				//msg = replaceXMLField(msg, "PackNo", GeneratePackNo)

				// 获取国库代码
				treCode, _ := getXMLFieldValue(msg, "DrawBackTreCode")

				// 发送7221报文给mq
				client.SendTipsMsg(msg, treCode)

				atomic.AddUint64(&staticsData.SedMsg7221Count, 1)
			}
			if msgNo == "7211" {
				// 【替换操作提前到转换的时候了，这里不需要再替换了】
				//now := time.Now()
				//// 替换MsgID
				//msg = replaceXMLField(msg, "MsgID", GenerateUniqueTipsId)
				//// 替换MsgRef
				//msg = replaceXMLField(msg, "MsgRef", GenerateUniqueTipsId)
				//// 替换DES
				//msg = replaceXMLFieldWithValue(msg, "DES", "333333333333")
				//// 替换PackNo
				//msg = replaceXMLField(msg, "PackNo", GeneratePackNo)
				//// 替换EntrustDate
				//msg = replaceXMLFieldWithValue(msg, "EntrustDate", now.Format("20060102"))
				//// 替换TraNo
				//msg = replaceXMLField(msg, "TraNo", GenerateTraNo)
				//// 替换TaxVouNo
				//msg = replaceXMLField(msg, "TaxVouNo", GenerateTaxVouNo)
				//// 替换BillDate
				//msg = replaceXMLFieldWithValue(msg, "BillDate", now.Format("20060102"))

				// 获取国库代码
				treCode, _ := getXMLFieldValue(msg, "PayeeTreCode")

				// 发送7211报文
				client.SendTipsMsg(msg, treCode)

				atomic.AddUint64(&staticsData.SedMsg7211Count, 1)
			}
		}
		return nil
	})
	if err != nil {
		AppLogger.Printf("处理文件失败: %v", err)
		return
	}
	AppLogger.Printf("Worker %d finished!\n", id)
}

func processSingleXMLFile(filePath string) (string, error) {
	// 读取 XML 文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		AppLogger.Printf("Error reading XML file %s: %v", filePath, err)
		return "", err
	}

	// 使用GBK解码器将字节数据解码为UTF-8字符串
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Data, err := decoder.Bytes(data)
	if err != nil {
		return "", err
	}

	return string(utf8Data), nil
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

func displayStatistics(ctx context.Context, data *StatisticsData, list *tview.TextView, app *tview.Application) {
	data.SedMsg7211Count = 0
	data.SedMsg7221Count = 0
	data.DecryptFileCount = 0
	data.ConvertFileCount = 0
	data.Detail7211Count = 0
	data.Detail7221Count = 0
	for {
		select {
		case <-ctx.Done():
			app.QueueUpdateDraw(func() {
				list.Clear()
				fmt.Fprintf(list, "解密文件数 [%d]\n", data.DecryptFileCount)
				fmt.Fprintf(list, "转换文件数 [%d]\n", data.ConvertFileCount)
				fmt.Fprintf(list, "发送7211报文数 [%d]\n", data.SedMsg7211Count)
				fmt.Fprintf(list, "发送7221报文数 [%d]\n", data.SedMsg7221Count)
				fmt.Fprintf(list, "7211明细数 [%d]\n", data.Detail7211Count)
				fmt.Fprintf(list, "7221明细数 [%d]\n", data.Detail7221Count)

				fmt.Fprintf(list, "\nCurrent Time is %s\n", time.Now().Format("2006-01-02T15:04:05"))
				fmt.Fprintf(list, "执行结束...")
			})
			AppLogger.Printf("Display worker stopping.\n")
			return
		default:
			app.QueueUpdateDraw(func() {
				list.Clear()
				fmt.Fprintf(list, "解密文件数 [%d]\n", data.DecryptFileCount)
				fmt.Fprintf(list, "转换文件数 [%d]\n", data.ConvertFileCount)
				fmt.Fprintf(list, "发送7211报文数 [%d]\n", data.SedMsg7211Count)
				fmt.Fprintf(list, "发送7221报文数 [%d]\n", data.SedMsg7221Count)

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

	// 创建页面管理器
	pages := tview.NewPages()

	form := tview.NewForm().
		AddInputField("cfmq地址", setting.Server, 50, nil, func(text string) { setting.Server = text }).
		AddInputField("用户名", setting.Username, 50, nil, func(text string) { setting.Username = text }).
		AddPasswordField("密码", setting.Password, 50, '*', func(text string) { setting.Password = text }).
		AddInputField("tips报文队列名", setting.SedQueueTips, 50, nil, func(text string) { setting.SedQueueTips = text }).
		AddInputField("原始文件路径", setting.FilePath, 50, nil, func(text string) { setting.FilePath = text }).
		AddButton("浏览...", func() {
			// 调用系统文件选择对话框
			go func() {
				dir, err := dialog.Directory().Title("Choose XML Directory").Browse()
				if err == nil {
					app.QueueUpdateDraw(func() {
						filePathField := globalFrom.GetFormItemByLabel("原始文件路径").(*tview.InputField)
						filePathField.SetText(dir)
						setting.FilePath = dir
					})
				}
			}()
		}).
		AddInputField("解密密钥", setting.EncKey, 50, nil, func(text string) { setting.EncKey = text }).
		AddInputField("退库报文银行行号", setting.PayeeOpBkCode, 50, nil, func(text string) { setting.PayeeOpBkCode = text }).
		//AddButton("Browse...", func() {
		//	showDirectoryBrowser(app, nil, setting)
		//}).
		AddButton(FireButtonName, func() {
			if globalFrom == nil {
				return
			}
			// 立即更新按钮状态为"运行中"
			button := globalFrom.GetButton(globalFrom.GetButtonIndex(FireButtonName))
			button.SetLabel("运行中")
			button.SetDisabled(true) // 按钮会置灰并禁用
			// 创建新的 context
			ctx, cancel = context.WithCancel(context.Background())
			go displayStatistics(ctx, statisticdata, statisticsList, app)
			go func() {
				handleMsg(ctx, 1, setting, statisticdata)
				cancel()
				app.QueueUpdateDraw(func() {
					button.SetLabel(FireButtonName)
					// 恢复正常状态
					button.SetDisabled(false)
				})
			}()
		}).
		AddButton("Quit", func() {
			setting.Save()
			if testClient != nil {
				testClient.Logout()
			}
			cancel()
			app.Stop()
		}).
		AddButton("Go to New Tab", func() {
			pages.SwitchToPage("newTab")
		})
	form.SetBorder(true).SetTitle("原始报文转换后推送ctbs").SetTitleAlign(tview.AlignCenter)
	globalFrom = form

	// 创建新的 tab 页面
	newForm1 := tview.NewForm().
		AddInputField("原始文件路径", setting.OriginalFilePath, 50, nil, func(text string) { setting.OriginalFilePath = text }).
		AddInputField("密钥", setting.EncKey, 50, nil, func(text string) { setting.EncKey = text }).
		AddButton(Decrypt, func() {
			if decryptForm == nil {
				return
			}
			// 立即更新按钮状态为"运行中"
			button := decryptForm.GetButton(decryptForm.GetButtonIndex(Decrypt))
			button.SetLabel("Running...")
			button.SetDisabled(true) // 按钮会置灰并禁用

			// 解密
			decryptFiles(setting.OriginalFilePath, setting.EncKey, statisticdata)

			// 更新按钮状态
			button.SetLabel(Decrypt)
			button.SetDisabled(false) // 恢复正常状态
		})
	newForm1.SetBorder(true).SetTitle("解密").SetTitleAlign(tview.AlignCenter)
	decryptForm = newForm1

	newForm2 := tview.NewForm().
		AddInputField("解密后文件路径", setting.DecryptedFilePath, 50, nil, func(text string) { setting.DecryptedFilePath = text }).
		AddInputField("退库报文银行行号", setting.PayeeOpBkCode, 50, nil, func(text string) { setting.PayeeOpBkCode = text }).
		AddButton(Convert, func() {
			if convertForm == nil {
				return
			}
			// 立即更新按钮状态为"运行中"
			button := convertForm.GetButton(convertForm.GetButtonIndex(Convert))
			button.SetLabel("Running...")
			button.SetDisabled(true) // 按钮会置灰并禁用

			// 转换
			convertFiles(setting.DecryptedFilePath, setting.OriginalFilePath, setting.PayeeOpBkCode, statisticdata)

			// 更新按钮状态
			button.SetLabel(Convert)
			button.SetDisabled(false) // 恢复正常状态
		}).
		AddButton("Back to Main", func() {
			pages.SwitchToPage("main")
		})
	newForm2.SetBorder(true).SetTitle("转换").SetTitleAlign(tview.AlignCenter)
	convertForm = newForm2

	// 创建页面布局
	flex := tview.NewFlex().
		AddItem(statisticsList, 0, 1, false).
		AddItem(form, 0, 1, true)
	// 创建左边的垂直布局，包含 newForm1 和 newForm2
	rightFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(newForm1, 0, 1, true).
		AddItem(newForm2, 0, 1, true)
	// 创建主水平布局，左边是垂直布局，右边是 statisticsList
	newFlex := tview.NewFlex().
		AddItem(statisticsList, 0, 1, false).
		AddItem(rightFlex, 0, 1, true)

	// 然后修改 Browse 按钮 - 需要重新获取 form 并修改按钮
	//form.GetButton(form.GetButtonIndex("Browse...")).SetSelectedFunc(func() {
	//	showDirectoryBrowser(app, flex, setting)
	//})

	// 添加页面到页面管理器
	pages.AddPage("main", flex, true, true)
	pages.AddPage("newTab", newFlex, true, false)

	// 运行应用
	//if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
	//	panic(err)
	//}
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
