package main

import (
	"encoding/xml"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// CFX 结构体定义XML根元素
type CFX struct {
	XMLName xml.Name `xml:"CFX"`
	HEAD    HEAD     `xml:"HEAD"`
	MSG     MSG      `xml:"MSG"`
}

// HEAD 结构体定义头部信息
type HEAD struct {
	VER      string `xml:"VER"`
	SRC      string `xml:"SRC"`
	DES      string `xml:"DES"`
	APP      string `xml:"APP"`
	MsgNo    string `xml:"MsgNo"`
	MsgID    string `xml:"MsgID"`
	MsgRef   string `xml:"MsgRef"`
	WorkDate string `xml:"WorkDate"`
	Reserve  string `xml:"Reserve,omitempty"`
}

// MSG 结构体定义消息体
type MSG struct {
	// 7221 消息类型相关字段
	BatchHead7221    *BatchHead7221    `xml:"BatchHead7221,omitempty"`
	DrawbackBody7221 *DrawbackBody7221 `xml:"DrawbackBody7221,omitempty"`

	// 6100 消息类型相关字段
	TaxHead6100 *TaxHead6100 `xml:"TaxHead6100,omitempty"`
	TaxBody6100 *TaxBody6100 `xml:"TaxBody6100,omitempty"`

	// 7211 消息类型相关字段
	BatchHead7211   *BatchHead7211   `xml:"BatchHead7211,omitempty"`
	TurnAccount7211 *TurnAccount7211 `xml:"TurnAccount7211,omitempty"`
	TaxBody7211     *TaxBody7211     `xml:"TaxBody7211,omitempty"`
}

// BatchHead7221 7221 消息类型的结构体
type BatchHead7221 struct {
	TaxOrgCode      string `xml:"TaxOrgCode"`
	EntrustDate     string `xml:"EntrustDate"`
	PackNo          string `xml:"PackNo"`
	DrawBackTreCode string `xml:"DrawBackTreCode"`
	ReckStyle       string `xml:"ReckStyle"`
	AllNum          string `xml:"AllNum"`
	AllAmt          string `xml:"AllAmt"`
}

type DrawbackBody7221 struct {
	DrawbackInfo7221 []DrawbackInfo7221 `xml:"DrawbackInfo7221"`
}

type DrawbackInfo7221 struct {
	TraNo              string `xml:"TraNo"`
	BillDate           string `xml:"BillDate"`
	VouNo              string `xml:"VouNo"`
	Amt                string `xml:"Amt"`
	BdgLevel           string `xml:"BdgLevel"`
	BdgKind            string `xml:"BdgKind"`
	BdgSbtCode         string `xml:"BdgSbtCode"`
	ViceSign           string `xml:"ViceSign"`
	TrimSign           string `xml:"TrimSign"`
	DrawBackReasonCode string `xml:"DrawBackReasonCode"`
	ApproveOrg         string `xml:"ApproveOrg"`
	PayeeOrgCode       string `xml:"PayeeOrgCode"`
	PayeeName          string `xml:"PayeeName"`
	TaxPayName         string `xml:"TaxPayName"`
	PayeeBankNo        string `xml:"PayeeBankNo"`
	PayeeOpBkCode      string `xml:"PayeeOpBkCode"`
	PayeeAcct          string `xml:"PayeeAcct"`
}

// TaxHead6100 6100 消息类型的结构体
type TaxHead6100 struct {
	ExportDate string `xml:"ExportDate"`
	ExportOrd  string `xml:"ExportOrd"`
	PayBkCode  string `xml:"PayBkCode"`
	TreCode    string `xml:"TreCode"`
}

type TaxBody6100 struct {
	TaxBill6100 []TaxBill6100 `xml:"TaxBill6100"`
}

type TaxBill6100 struct {
	TaxOrgCode        string `xml:"TaxOrgCode"`
	TaxAmt            string `xml:"TaxAmt"`
	ExpTaxVouNo       string `xml:"ExpTaxVouNo"`
	BudgetType        string `xml:"BudgetType"`
	TrimSign          string `xml:"TrimSign"`
	BudgetSubjectCode string `xml:"BudgetSubjectCode"`
	BudgetLevelCode   string `xml:"BudgetLevelCode"`
	ViceSign          string `xml:"ViceSign"`
}

// BatchHead7211 7211 消息类型的结构体
type BatchHead7211 struct {
	TaxOrgCode  string `xml:"TaxOrgCode"`
	EntrustDate string `xml:"EntrustDate"`
	PackNo      string `xml:"PackNo"`
	AllNum      int    `xml:"AllNum"`
	AllAmt      string `xml:"AllAmt"`
}

type TurnAccount7211 struct {
	BizType      int    `xml:"BizType"`
	FundSrlNo    string `xml:"FundSrlNo"`
	PayBnkNo     string `xml:"PayBnkNo"`
	PayeeTreCode string `xml:"PayeeTreCode"`
	PayeeTreName string `xml:"PayeeTreName"`
}

type TaxBody7211 struct {
	TaxInfo7211 []TaxInfo7211 `xml:"TaxInfo7211"`
}

type TaxInfo7211 struct {
	Payment7211 Payment7211 `xml:"Payment7211"`
	TaxVou7211  TaxVou7211  `xml:"TaxVou7211"`
	TaxType7211 TaxType7211 `xml:"TaxType7211"`
}

type Payment7211 struct {
	TraNo        string `xml:"TraNo"`
	TraAmt       string `xml:"TraAmt"`
	PayOpBnkNo   string `xml:"PayOpBnkNo"`
	PayOpBnkName string `xml:"PayOpBnkName"`
	HandOrgName  string `xml:"HandOrgName"`
	PayAcct      string `xml:"PayAcct"`
}

type TaxVou7211 struct {
	TaxVouNo   string `xml:"TaxVouNo"`
	BillDate   string `xml:"BillDate"`
	TaxPayCode string `xml:"TaxPayCode"`
	TaxPayName string `xml:"TaxPayName"`
	BudgetType string `xml:"BudgetType"`
	TrimSign   string `xml:"TrimSign"`
}

type TaxType7211 struct {
	BudgetSubjectCode string            `xml:"BudgetSubjectCode"`
	LimitDate         string            `xml:"LimitDate"`
	BudgetLevelCode   string            `xml:"BudgetLevelCode"`
	BudgetLevelName   string            `xml:"BudgetLevelName"`
	TaxStartDate      string            `xml:"TaxStartDate"`
	TaxEndDate        string            `xml:"TaxEndDate"`
	ViceSign          string            `xml:"ViceSign"`
	TaxType           int               `xml:"TaxType"`
	HandBookKind      int               `xml:"HandBookKind"`
	DetailNum         int               `xml:"DetailNum"`
	SubjectList7211   []SubjectList7211 `xml:"SubjectList7211"`
}

type SubjectList7211 struct {
	DetailNo       int    `xml:"DetailNo"`
	TaxSubjectCode string `xml:"TaxSubjectCode"`
	TaxSubjectName string `xml:"TaxSubjectName"`
	TaxNumber      string `xml:"TaxNumber"`
	TaxAmt         string `xml:"TaxAmt"`
	TaxRate        string `xml:"TaxRate"`
	ExpTaxAmt      string `xml:"ExpTaxAmt"`
	DiscountTaxAmt string `xml:"DiscountTaxAmt"`
	FactTaxAmt     string `xml:"FactTaxAmt"`
}

// OutputData 结构体定义输出数据
type OutputData struct {
	Item1           string
	Item2           string
	Item3           string
	Item4           string
	Item5           string
	Item6           string
	Item7           string
	Item8           string
	Detail7211Count int
	Detail7221Count int
}

// 全局变量
var (
	payeeOpBkCode string
	totalCounts   int
)

// ConvertMsg 读取并转换消息文件
func ConvertMsg(filePath string, payeeOpBkCodeNew string) ([]*OutputData, error) {
	// 检查文件名是否包含特定字符串
	if strings.Contains(filePath, "_3190_") {
		return nil, nil
	}

	// 读取文件内容（使用GBK编码）
	gbkDecoder := simplifiedchinese.GBK.NewDecoder()
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 将GBK编码转换为UTF-8
	utf8Content, err := gbkDecoder.Bytes(fileContent)
	if err != nil {
		return nil, fmt.Errorf("GBK解码失败: %v", err)
	}

	// 在 xml.Unmarshal 之前添加以下代码来移除编码声明
	xmlContent := string(utf8Content)
	// 移除GBK编码声明，替换为UTF-8
	xmlContent = strings.Replace(xmlContent, `encoding="GBK"`, `encoding="UTF-8"`, -1)
	xmlContent = strings.Replace(xmlContent, `encoding="gbk"`, `encoding="UTF-8"`, -1)
	xmlContent = strings.Replace(xmlContent, `encoding='GBK'`, `encoding='UTF-8'`, -1)
	xmlContent = strings.Replace(xmlContent, `encoding='gbk'`, `encoding='UTF-8'`, -1)

	// 解析XML
	var cfx CFX
	err = xml.Unmarshal([]byte(xmlContent), &cfx)
	if err != nil {
		return nil, fmt.Errorf("解析XML失败: %v", err)
	}

	payeeOpBkCode = payeeOpBkCodeNew
	// 根据消息类型调用相应的转换函数
	switch GetMsgType(&cfx) {
	case "7221":
		return Convert7221(&cfx)
	case "6100":
		return Convert6100To7211(&cfx)
	default:
		return nil, nil
	}
}

// GetMsgType 获取消息类型
func GetMsgType(cfx *CFX) string {
	// 检查是否为6100消息类型
	if cfx.MSG.TaxHead6100 != nil {
		return "6100"
	}

	// 检查HEAD中的MsgNo字段
	if cfx.HEAD.MsgNo == "" {
		return ""
	}

	return cfx.HEAD.MsgNo
}

// Convert7221 转换7221消息类型
func Convert7221(cfx *CFX) ([]*OutputData, error) {
	// 检查数量限制
	allNum, _ := strconv.Atoi(cfx.MSG.BatchHead7221.AllNum)
	if allNum > 1000 || len(cfx.MSG.DrawbackBody7221.DrawbackInfo7221) > 1000 {
		return nil, fmt.Errorf("7221 all count > 1000, the app should be update the logic")
	}

	msgid := GenerateUniqueTipsId()

	data := &OutputData{}
	data.Item4 = "7221"
	data.Item5 = BuildItem5(data.Item4, msgid)
	data.Item6 = BuildItem6()
	data.Item7 = BuildItem7(data.Item4)

	detail7221Count := 0

	// 构建新的CFX结构
	newCfx := CFX{
		HEAD: HEAD{
			VER:      cfx.HEAD.VER,
			SRC:      cfx.HEAD.SRC,
			DES:      "333333333333",
			APP:      cfx.HEAD.APP,
			MsgNo:    cfx.HEAD.MsgNo,
			MsgID:    msgid,
			MsgRef:   msgid,
			WorkDate: time.Now().Format("20060102"),
		},
		MSG: MSG{
			BatchHead7221: &BatchHead7221{
				TaxOrgCode:      cfx.MSG.BatchHead7221.TaxOrgCode,
				EntrustDate:     cfx.MSG.BatchHead7221.EntrustDate,
				PackNo:          GeneratePackNo(),
				DrawBackTreCode: cfx.MSG.BatchHead7221.DrawBackTreCode,
				ReckStyle:       cfx.MSG.BatchHead7221.ReckStyle,
				AllNum:          cfx.MSG.BatchHead7221.AllNum,
				AllAmt:          cfx.MSG.BatchHead7221.AllAmt,
			},
		},
	}

	// 创建DrawbackBody7221
	drawbackBody := &DrawbackBody7221{}
	for _, info := range cfx.MSG.DrawbackBody7221.DrawbackInfo7221 {
		drawbackInfo := DrawbackInfo7221{
			TraNo:              info.TraNo,
			BillDate:           info.BillDate,
			VouNo:              info.VouNo,
			Amt:                info.Amt,
			BdgLevel:           info.BdgLevel,
			BdgKind:            info.BdgKind,
			BdgSbtCode:         info.BdgSbtCode,
			ViceSign:           info.ViceSign,
			TrimSign:           info.TrimSign,
			DrawBackReasonCode: info.DrawBackReasonCode,
			ApproveOrg:         info.ApproveOrg,
			PayeeOrgCode:       info.PayeeOrgCode,
			PayeeName:          info.PayeeName,
			TaxPayName:         info.TaxPayName,
			PayeeBankNo:        info.PayeeBankNo,
			PayeeOpBkCode:      payeeOpBkCode,
			PayeeAcct:          info.PayeeAcct,
		}

		// 如果PayeeOpBkCode为空，则使用原值
		if payeeOpBkCode == "" {
			drawbackInfo.PayeeOpBkCode = info.PayeeOpBkCode
		}

		drawbackBody.DrawbackInfo7221 = append(drawbackBody.DrawbackInfo7221, drawbackInfo)
		totalCounts++
		detail7221Count++
	}

	newCfx.MSG.DrawbackBody7221 = drawbackBody

	// 生成XML字符串
	outputXML, err := generateXMLString(newCfx)
	if err != nil {
		return nil, err
	}

	data.Item8 = outputXML
	data.Detail7221Count = detail7221Count
	result := []*OutputData{data}
	return result, nil
}

// Convert6100To7211 转换6100到7211消息类型
func Convert6100To7211(cfx *CFX) ([]*OutputData, error) {
	var result []*OutputData

	// 按TaxOrgCode分组处理
	taxOrgGroups := make(map[string][]TaxBill6100)
	for _, bill := range cfx.MSG.TaxBody6100.TaxBill6100 {
		taxOrgGroups[bill.TaxOrgCode] = append(taxOrgGroups[bill.TaxOrgCode], bill)
	}

	// 如果有多个TaxOrgCode分组，分别处理
	if len(taxOrgGroups) > 1 {
		for taxOrgCode, bills := range taxOrgGroups {
			// 创建新的CFX结构
			newCfx := CFX{
				MSG: MSG{
					TaxHead6100: cfx.MSG.TaxHead6100,
					TaxBody6100: &TaxBody6100{
						TaxBill6100: bills,
					},
				},
			}

			// 处理单个分组
			groupResults := doConvert6100To7211(&newCfx, taxOrgCode)
			result = append(result, groupResults...)
		}
		return result, nil
	} else {
		// 单个分组处理
		return doConvert6100To7211(cfx, ""), nil
	}
}

// doConvert6100To7211 实际执行6100到7211的转换
func doConvert6100To7211(cfx *CFX, taxOrgCode string) []*OutputData {
	result := []*OutputData{}
	j := 1
	n := 1

	taxBody7211 := &TaxBody7211{}
	total6100Detail := len(cfx.MSG.TaxBody6100.TaxBill6100)
	var totalAmt float64

	detail7211Count := 0
	for _, body6100 := range cfx.MSG.TaxBody6100.TaxBill6100 {
		amt, _ := strconv.ParseFloat(body6100.TaxAmt, 64)
		totalAmt += amt

		taxInfo7211 := TaxInfo7211{
			Payment7211: Payment7211{
				TraNo:        GenerateTraNo(),
				TraAmt:       body6100.TaxAmt,
				PayOpBnkNo:   "319376074422",
				PayOpBnkName: "付款开户行名称",
				HandOrgName:  "缴款单位名称",
				PayAcct:      "9090313010010991008844",
			},
			TaxVou7211: TaxVou7211{
				TaxVouNo:   GenerateTaxVouNo(),
				BillDate:   time.Now().Format("20060102"),
				TaxPayCode: "123508214908270504",
				TaxPayName: "长汀县新桥中学",
				BudgetType: body6100.BudgetType,
				TrimSign:   body6100.TrimSign,
			},
			TaxType7211: TaxType7211{
				BudgetSubjectCode: body6100.BudgetSubjectCode,
				LimitDate:         time.Now().Format("20060102"),
				BudgetLevelCode:   body6100.BudgetLevelCode,
				BudgetLevelName:   "省",
				TaxStartDate:      time.Now().Format("20060102"),
				TaxEndDate:        time.Now().Format("20060102"),
				ViceSign:          body6100.ViceSign,
				TaxType:           3,
				HandBookKind:      3,
				DetailNum:         1,
				SubjectList7211: []SubjectList7211{
					{
						DetailNo:       1,
						TaxSubjectCode: "11111111111111111111",
						TaxSubjectName: "税目111",
						TaxNumber:      "1",
						TaxAmt:         body6100.TaxAmt,
						TaxRate:        "1.00",
						ExpTaxAmt:      body6100.TaxAmt,
						DiscountTaxAmt: body6100.TaxAmt,
						FactTaxAmt:     body6100.TaxAmt,
					},
				},
			},
		}

		taxBody7211.TaxInfo7211 = append(taxBody7211.TaxInfo7211, taxInfo7211)
		totalCounts++
		detail7211Count++

		// 每1000条记录或最后一条记录时生成一个输出
		if j%1000 == 0 || total6100Detail == j {
			msgId := GenerateUniqueTipsId()

			data := &OutputData{}
			data.Item4 = "7211"
			data.Item5 = BuildItem5(data.Item4, msgId)
			data.Item6 = BuildItem6()
			data.Item7 = BuildItem7(data.Item4)
			data.Detail7211Count = detail7211Count

			// 构建新的CFX结构
			newCfx := CFX{
				HEAD: HEAD{
					VER:      "1.0",
					SRC:      "100000000000",
					DES:      "333333333333",
					APP:      "TIPS",
					MsgNo:    "7211",
					MsgID:    msgId,
					MsgRef:   msgId,
					WorkDate: time.Now().Format("20060102"),
					Reserve:  "预留字段预留字段预留字段预留字段预留字段",
				},
				MSG: MSG{
					BatchHead7211: &BatchHead7211{
						TaxOrgCode:  getFirstTaxOrgCode(cfx),
						EntrustDate: time.Now().Format("20060102"),
						PackNo:      GeneratePackNo(),
						AllNum:      n,
						AllAmt:      fmt.Sprintf("%.2f", totalAmt),
					},
					TurnAccount7211: &TurnAccount7211{
						BizType:      0,
						FundSrlNo:    GenerateUniqueId(),
						PayBnkNo:     cfx.MSG.TaxHead6100.PayBkCode,
						PayeeTreCode: cfx.MSG.TaxHead6100.TreCode,
						PayeeTreName: "收款国库名称",
					},
					TaxBody7211: taxBody7211,
				},
			}

			// 生成XML字符串
			outputXML, err := generateXMLString(newCfx)
			if err == nil {
				data.Item8 = outputXML
				result = append(result, data)
			}

			// 重置计数器和累计金额
			taxBody7211 = &TaxBody7211{}
			totalAmt = 0
			n = 0
			detail7211Count = 0
		}

		j++
		n++
	}

	return result
}

// getFirstTaxOrgCode 获取第一个TaxOrgCode
func getFirstTaxOrgCode(cfx *CFX) string {
	if len(cfx.MSG.TaxBody6100.TaxBill6100) > 0 {
		return cfx.MSG.TaxBody6100.TaxBill6100[0].TaxOrgCode
	}
	return ""
}

// generateXMLString 生成格式化的XML字符串
func generateXMLString(cfx CFX) (string, error) {
	output, err := xml.MarshalIndent(cfx, "", "")
	if err != nil {
		return "", err
	}

	// 转换为GBK编码
	gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	gbkOutput, err := gbkEncoder.String(string(output))
	if err != nil {
		return "", err
	}

	// 移除换行符和空格并添加XML声明
	cleanedOutput := strings.ReplaceAll(gbkOutput, "\n", "")
	cleanedOutput = strings.ReplaceAll(cleanedOutput, "\r", "")
	cleanedOutput = strings.ReplaceAll(cleanedOutput, " ", "")

	return `<?xml version="1.0" encoding="GBK"?>` + cleanedOutput, nil
}

// BuildItem5 构建Item5字段
func BuildItem5(msgType, msgId string) string {
	return fmt.Sprintf("{\"SRC\":\"100000000000\",\"DES\":\"111111111111\",\"MsgNo\":\"%s\",\"MsgID\":\"%s\",\"MsgRef\":\"%s\",\"OriMsgNo\":\"\"}",
		msgType, msgId, msgId)
}

// BuildItem6 构建Item6字段
func BuildItem6() string {
	return "{\"SourceQueue\":\"TIPS.GK200.INT.NORMAL.IN.TRANSLOG\",\"Persistence\":\"1\",\"CORRELID\":\"000000000000000000000000000000000000000000000000\",\"MSGID\":\"414D5120514D5F544950535F494E545FAD50C9664C80DB20\",\"UserIdentifier\":\"mqm\",\"Priority\":\"4\",\"Encoding\":\"273\",\"CodedCharSetId\":\"819\",\"Expiry\":\"-1\",\"PutDate\":\"2024.08.27\",\"PutTime\":\"14:36:20.130\",\"ReplyToQMgr\":\"QM_TIPS_INT_34\",\"ReplyToQ\":\"\",\"Report\":\"0\",\"Format\":\"MQHRF2\",\"ApplIdentityData\":\"\"}"
}

// BuildItem7 构建Item7字段
func BuildItem7(msgType string) string {
	return fmt.Sprintf("{\"Encoding\":\"273\",\"CodedCharSetId\":\"819\",\"Dlv\":\"UNKNOWN\",\"Pri\":\"UNKNOWN\",\"ISCURRCENTER\":\"UNKNOWN\",\"MSGFORMAT\":\"TIPS\",\"MSGFLAG\":\"NORMAL\",\"MSGORIFLAG\":\"UNKNOWN\",\"Flags\":\"0\",\"Format\":\"\",\"NameValueCCSID\":\"1208\",\"Version\":\"2\",\"APP\":\"TIPS\",\"CorrelationID\":\"524551000000000000000000000000000000000000000000\",\"DES\":\"111111111111\",\"DESTINATION\":\"GK200000002000\",\"EXCEPTIONMSG\":\"UNKNOWN\",\"EXCEPTIONTYPE\":\"UNKNOWN\",\"MSGDESC\":\"TIPS\",\"MSGTYPE\":\"MSG\",\"MsgID\":\"24082772110050119383\",\"MsgRef\":\"01610700000023848006\",\"ServiceId\":\"UNKNOWN\",\"ServiceId9120\":\"UNKNOWN\",\"TIPSTAGFLAG\":\"UNKNOWN\",\"TIPS_MSGNO\":\"UNKNOWN\",\"TIPS_ORIMSGNO\":\"UNKNOWN\",\"msgFormat\":\"TIPS\",\"msgReserve\":\"\",\"msgSrcAddr\":\"100000000000\",\"msgTarAddr\":\"111111111111\",\"msgType\":\"%s\",\"msgVer\":\"1.0\"}", msgType)
}
