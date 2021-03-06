package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/errno"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/httputils"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/dao"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/model"
	"github.com/gin-gonic/gin"
)

type auditItem struct {
	Address     string `json:"address" wx:"address"`
	Tag         string `json:"tag" wx:"tag"`
	FirstClass  string `json:"firstClass" wx:"first_class"`
	SecondClass string `json:"secondClass" wx:"second_class"`
	ThirdClass  string `json:"thirdClass" wx:"third_class"`
	FirstId     int    `json:"firstId" wx:"first_id"`
	SecondId    int    `json:"secondId" wx:"second_id"`
	ThirdId     int    `json:"thirdId" wx:"third_id"`
	Title       string `json:"title" wx:"title"`
}

type previewInfo struct {
	VideoIdList []string `json:"videoIdList" wx:"video_id_list"`
	PicIdList   []string `json:"picIdList" wx:"pic_id_list"`
}

type ugcDeclare struct {
	Scene          []int  `json:"scene" wx:"scene"`
	OtherSceneDesc string `json:"otherSceneDesc" wx:"other_scene_desc"`
	Method         []int  `json:"method" wx:"method"`
	HasAuditTeam   int    `json:"hasAuditTeam" wx:"has_audit_team"`
	AuditDesc      string `json:"auditDesc" wx:"audit_desc"`
}

type submitAuditReq struct {
	ItemList      []auditItem `json:"itemList" wx:"item_list"`
	PreviewInfo   previewInfo `json:"previewInfo" wx:"preview_info"`
	VersionDesc   string      `json:"versionDesc" wx:"version_desc"`
	FeedbackInfo  string      `json:"feedbackInfo" wx:"feedback_info"`
	FeedbackStuff string      `json:"feedbackStuff" wx:"feedback_stuff"`
	UgcDeclare    ugcDeclare  `json:"ugcDeclare" wx:"ugc_declare"`
}

type submitAuditResp struct {
	AuditId int `json:"auditId" wx:"auditid"`
}

type getLatestAuditStatusResp struct {
	AuditId         int64  `json:"auditId" wx:"auditid"`
	Status          int    `json:"status" wx:"status"`
	Reason          string `json:"reason" wx:"reason"`
	ScreenShot      string `json:"screenShot" wx:"ScreenShot"`
	UserVersion     string `json:"userVersion" wx:"user_version"`
	UserDesc        string `json:"userDesc" wx:"user_desc"`
	SubmitAuditTime int64  `json:"submitAuditTime" wx:"submit_audit_time"`
}

type devVersionsResp struct {
	AuditVersion *getLatestAuditStatusResp `json:"auditInfo,omitempty"`
	getVersionInfoResp
}

type templateListResp struct {
	TemplateList []templateItem `json:"templateList" wx:"template_list"`
}

type templateItem struct {
	CreateTime             int64          `json:"createTime" wx:"create_time"`
	UserVersion            string         `json:"userVersion" wx:"user_version"`
	UserDesc               string         `json:"userDesc" wx:"user_desc"`                              // ???????????????????????????????????????
	TemplateId             int            `json:"templateId" wx:"template_id"`                          // ?????? id
	TemplateType           int            `json:"templateType" wx:"template_type"`                      // 0?????????????????????1??????????????????
	SourceMiniprogramAppid string         `json:"sourceMiniprogramAppid" wx:"source_miniprogram_appid"` // ??????????????????appid
	SourceMiniprogram      string         `json:"sourceMiniprogram" wx:"source_miniprogram"`            // ????????????????????????
	CategoryList           []categoryItem `json:"categoryList" wx:"category_list"`                      // [???????????????????????????](#category_list????????????????????????)?????????????????????????????????????????????
	AuditScene             int            `json:"auditScene" wx:"audit_scene"`                          // ?????????????????????????????????????????????????????????
	AuditStatus            int            `json:"auditStatus" wx:"audit_status"`                        // ?????????????????????????????????????????????????????????
	Reason                 string         `json:"reason" wx:"reason"`                                   // ?????????????????????????????????????????????????????????????????????
}

type categoryItem struct {
	FirstClass  string `json:"firstClass" wx:"first_class"`   // ????????????
	FirstId     int    `json:"firstId" wx:"first_id"`         // ????????????id
	SecondClass string `json:"secondClass" wx:"second_class"` // ????????????
	SecondId    int    `json:"secondId" wx:"second_id"`       // ????????????id
}

type codeCommitReq struct {
	TemplateId  string `json:"templateId" wx:"template_id"`   // ??????????????????????????? ID????????????[????????????????????????](https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/code_template/gettemplatelist.html)????????????template_id <br>????????????????????????id???????????????????????????id??????ext_json????????????????????????{"extAppid":" ", "ext": {}, "window": {}}
	ExtJson     string `json:"extJson" wx:"ext_json"`         // ????????????????????????????????????????????? extAppid ??????????????????????????????[ext.json????????????](https://developers.weixin.qq.com/miniprogram/dev/devtools/ext.html#%E5%B0%8F%E7%A8%8B%E5%BA%8F%E6%A8%A1%E6%9D%BF%E5%BC%80%E5%8F%91)????????????????????????????????????ext.json????????????????????????????????????????????????????????????????????????"ext_json????????????"???
	UserVersion string `json:"userVersion" wx:"user_version"` // ???????????????????????????????????????????????????????????? 64 ????????????
	UserDesc    string `json:"userDesc" wx:"user_desc"`       // ????????????????????????????????????
}

type visitStatusResp struct {
	Status int `wx:"status"`
}

type releaseInfo struct {
	ReleaseTime    int64  `json:"releaseTime" wx:"release_time"`
	ReleaseVersion string `json:"releaseVersion" wx:"release_version"`
	ReleaseDesc    string `json:"releaseDesc" wx:"release_desc"`
	ReleaseQrCode  string `json:"releaseQrCode,omitempty"`
}

type expInfo struct {
	ExpTime    int64  `json:"expTime" wx:"exp_time"`
	ExpVersion string `json:"expVersion" wx:"exp_version"`
	ExpDesc    string `json:"expDesc" wx:"exp_desc"`
	ExpQrCode  string `json:"expQrCode,omitempty"`
}

type getVersionInfoResp struct {
	ReleaseInfo *releaseInfo `json:"releaseInfo,omitempty" wx:"release_info"`
	ExpInfo     *expInfo     `json:"expInfo,omitempty" wx:"exp_info"`
}
type getDevWeAppListResp struct {
	Appid         string `json:"appid"`
	NickName      string `json:"nickName"`
	FuncInfo      []int  `json:"funcInfo"`
	QrCodeUrl     string `json:"qrCodeUrl"`
	ServiceStatus int    `json:"serviceStatus"`
	getVersionInfoResp
}

type uploadMediaResp struct {
	Type      string `json:"type" wx:"type"`
	MediaId   string `json:"mediaId" wx:"media_id"`
	CreatedAt int64  `json:"createdAt" wx:"created_at"`
}

type changeVisitStatusReq struct {
	Action string `json:"action"`
}

type pageList struct {
	PageList []string `json:"pageList" wx:"page_list"`
}

type category struct {
	FirstClass  string `json:"firstClass" wx:"first_class"`
	SecondClass string `json:"secondClass" wx:"second_class"`
	ThirdClass  string `json:"thirdClass" wx:"third_class"`
	FirstId     int    `json:"firstId" wx:"first_id"`
	SecondId    int    `json:"secondId" wx:"second_id"`
	ThirdId     int    `json:"thirdId" wx:"third_id"`
}
type categoryList struct {
	CategoryList []category `json:"categoryList" wx:"category_list"`
}

func submitAudit(appid string, req *submitAuditReq) (int, error) {
	_, body, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/submit_audit", "", *req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	var resp submitAuditResp
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return 0, err
	}
	return resp.AuditId, nil
}

func getLatestAuditStatus(appid string, resp *getLatestAuditStatusResp) (bool, error) {
	wxerr, body, err := wx.GetWxApiWithAuthToken(appid, "/wxa/get_latest_auditstatus", "")
	if err != nil {
		if wxerr != nil && wxerr.ErrCode == 85058 {
			return false, nil
		}
		return false, err
	}
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return false, err
	}
	return true, nil
}

func getVisitStatus(appid string) (int, error) {
	_, body, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/getvisitstatus", "", gin.H{})
	if err != nil {
		log.Error(err)
		return 0, err
	}
	var resp visitStatusResp
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return 0, err
	}
	return resp.Status, nil
}

func getVersionInfo(appid string, resp *getVersionInfoResp) error {
	_, body, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/getversioninfo", "", gin.H{})
	if err != nil {
		log.Error(err)
		return err
	}
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return err
	}
	return nil
}

func getImageResp(resp *http.Response, body []byte) (string, error) {
	if len(resp.Header["Content-Type"]) > 0 && resp.Header["Content-Type"][0] == "image/jpeg" {
		return base64.StdEncoding.EncodeToString(body), nil
	}
	var wxError wx.WxCommError
	if err := wx.WxJson.Unmarshal(body, &wxError); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return "", err
	}
	if wxError.ErrCode != 0 {
		return "", fmt.Errorf("WxErrCode != 0, resp: %v", wxError)
	}
	return "", fmt.Errorf("unknown error, resp: %v", body)
}

func getReleaseQrCode(appid string) (string, error) {
	url, err := wx.GetAuthorizerWxApiUrl(appid, "/wxa/getwxacodeunlimit", "")
	if err != nil {
		log.Error(err)
		return "", err
	}
	jsonByte, _ := json.Marshal(gin.H{"scene": "wxcomponent"})
	resp, body, err := httputils.RawPost(url, jsonByte, "application/json")
	if err != nil {
		log.Error(err)
		return "", err
	}
	return getImageResp(resp, body)
}

func getExpQrCode(appid string) (string, error) {
	url, err := wx.GetAuthorizerWxApiUrl(appid, "/wxa/get_qrcode", "")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resp, body, err := httputils.RawGet(url)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return getImageResp(resp, body)
}

func getDevWeAppListHandler(c *gin.Context) {
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	count, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	if count > 20 {
		c.JSON(http.StatusOK, errno.ErrInvalidParam)
		return
	}
	appid := c.DefaultQuery("appid", "")

	// ??????????????????
	records, total, err := dao.GetDevWeAppRecords(offset, count, appid)
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}

	// ????????????
	wg := &sync.WaitGroup{}
	wg.Add(len(records))
	resp := make([]getDevWeAppListResp, len(records))
	for i, record := range records {
		go func(i int, record *model.Authorizer) {
			defer wg.Done()
			resp[i].Appid = record.Appid
			resp[i].NickName = record.NickName
			resp[i].QrCodeUrl = record.QrcodeUrl

			// ?????????????????????
			strFuncInfoList := strings.Split(record.FuncInfo, "|")
			for _, v := range strFuncInfoList {
				id, err := strconv.Atoi(v)
				if err == nil {
					resp[i].FuncInfo = append(resp[i].FuncInfo, id)
				}
			}
			// ??????????????????
			status, err := getVisitStatus(record.Appid)
			if err != nil {
				log.Error(err)
			} else {
				resp[i].ServiceStatus = status
			}

			// ??????????????????
			var versionInfo getVersionInfoResp
			err = getVersionInfo(record.Appid, &versionInfo)
			if err != nil {
				log.Error(err)
			} else {
				resp[i].ReleaseInfo = versionInfo.ReleaseInfo
				resp[i].ExpInfo = versionInfo.ExpInfo
			}
		}(i, record)

	}
	wg.Wait()

	c.JSON(http.StatusOK, errno.OK.WithData(gin.H{"total": total, "records": resp}))
}

func submitAuditHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	var req submitAuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	auditId, err := submitAudit(appid, &req)
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(gin.H{"auditId": auditId}))
}

func devVersionsHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	var resp devVersionsResp
	var wg sync.WaitGroup
	wg.Add(1)
	// ????????????
	go func() {
		defer wg.Done()
		var auditInfo getLatestAuditStatusResp
		has, err := getLatestAuditStatus(appid, &auditInfo)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
			return
		}
		if has {
			resp.AuditVersion = &auditInfo
		}
	}()

	// ????????????????????????
	var versionInfo getVersionInfoResp
	err := getVersionInfo(appid, &versionInfo)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	if versionInfo.ExpInfo != nil {
		log.Info("get exp qrcode")
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp.ExpInfo = versionInfo.ExpInfo
			base64Image, err := getExpQrCode(appid)
			if err != nil {
				log.Error(err)
			} else {
				resp.ExpInfo.ExpQrCode = base64Image
			}
		}()
	}
	if versionInfo.ReleaseInfo != nil {
		log.Info("get release qrcode")
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp.ReleaseInfo = versionInfo.ReleaseInfo
			base64Image, err := getReleaseQrCode(appid)
			if err != nil {
				log.Error(err)
			} else {
				resp.ReleaseInfo.ReleaseQrCode = base64Image
			}
		}()
	}
	wg.Wait()
	c.JSON(http.StatusOK, errno.OK.WithData(resp))
}

func templateListHandler(c *gin.Context) {
	var resp templateListResp
	templateType := c.DefaultQuery("templateType", "")
	_, body, err := wx.GetWxApiWithComponentToken("/wxa/gettemplatelist", "template_type="+templateType)
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(resp))
}

func revokeAuditHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	_, _, err := wx.GetWxApiWithAuthToken(appid, "/wxa/undocodeaudit", "")
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func speedUpAuditHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	auditId, err := strconv.Atoi(c.DefaultQuery("auditId", "0"))
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	_, _, err = wx.PostWxJsonWithAuthToken(appid, "/wxa/speedupaudit", "", gin.H{"auditid": auditId})
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func commitCodeHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	var req codeCommitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	if _, _, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/commit", "", req); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func releaseCodeHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	if _, _, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/release", "", gin.H{}); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func uploadMediaHandler(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "")
	appid := c.DefaultQuery("appid", "")
	formFile, fileHeader, err := c.Request.FormFile("media")
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	_, body, err := wx.PostWxFormDataWithAuthToken(appid, "/cgi-bin/media/upload",
		"type="+mediaType, formFile, fileHeader.Filename, "media")
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	var resp uploadMediaResp
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(resp))
}

func changeVisitStatusHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	var req changeVisitStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	if _, _, err := wx.PostWxJsonWithAuthToken(appid, "/wxa/change_visitstatus",
		"", gin.H{"action": req.Action}); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func rollbackReleaseVersionHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	if _, _, err := wx.GetWxApiWithAuthToken(appid, "/wxa/revertcoderelease", ""); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}

func getPageListHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	_, body, err := wx.GetWxApiWithAuthToken(appid, "/wxa/get_page", "")
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	var resp pageList
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(resp))
}

func getCategoryHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	_, body, err := wx.GetWxApiWithAuthToken(appid, "/wxa/get_category", "")
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	var resp categoryList
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(resp))
}

func getQRCodeHandler(c *gin.Context) {
	appid := c.DefaultQuery("appid", "")
	base64Image, err := getReleaseQrCode(appid)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(gin.H{"releaseQrCode": base64Image}))
}
