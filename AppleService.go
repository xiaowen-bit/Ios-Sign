package appleservice

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"kksigncustom/models"
	"kksigncustom/service/ossservice"
	"kksigncustom/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	jsoniter "github.com/json-iterator/go"
	"howett.net/plist"
)

// 注册设备 post
const registdeviceURL string = "https://api.appstoreconnect.apple.com/v1/devices"

// 注册bundleid post
const registBundleidURL string = "https://api.appstoreconnect.apple.com/v1/bundleIds"

// registProfile 注册描述文件 post
const registProfileURL string = "https://api.appstoreconnect.apple.com/v1/profiles"

// AppleService Apple服务
type AppleService struct {
}

// NewAppleService 创建
func NewAppleService() *AppleService {
	return &AppleService{}
}

var reInfoPlist = regexp.MustCompile(`Payload/[^/]+/Info\.plist`)

// IosPlist ipa里面 info.plist 解析的结构
type IosPlist struct {
	CFBundleName         string `plist:"CFBundleName"`
	CFBundleDisplayName  string `plist:"CFBundleDisplayName"`
	CFBundleVersion      string `plist:"CFBundleVersion"`
	CFBundleShortVersion string `plist:"CFBundleShortVersionString"`
	CFBundleIdentifier   string `plist:"CFBundleIdentifier"`
	AppIcon              string
	MobileconfigPath     string
}

// Unzipipa 解压ipa文件
func (a *AppleService) Unzipipa(ipapath string) (ipainfo *IosPlist, err error) {
	file, err := os.Open(ipapath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, err
	}
	var plistFile, iosIconFile *zip.File
	for _, f := range reader.File {
		switch {
		case reInfoPlist.MatchString(f.Name):
			plistFile = f
			break
		case strings.Contains(f.Name, "AppIcon60x60"):
			iosIconFile = f
			break
		}
	}
	if plistFile == nil {
		return nil, errors.New("info.plist is not found")
	}
	rc, err := plistFile.Open()
	if err != nil {
		return nil, errors.New("打开info.plst失败")
	}
	defer rc.Close()
	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, errors.New("打开info.plst失败")
	}
	p := new(IosPlist)
	decoder := plist.NewDecoder(bytes.NewReader(buf))
	if err := decoder.Decode(p); err != nil {
		return nil, errors.New("解析info.plst失败")
	}
	if iosIconFile == nil {
		return nil, errors.New("Icon is not found")
	}
	rc, err = iosIconFile.Open()
	if err != nil {
		return nil, errors.New("Icon is not found")
	}
	defer rc.Close()
	//保存applogo图片
	appiconpath := utils.SplicingString(filepath.Dir(ipapath), "/appicon.png")
	newFile, err := os.Create(appiconpath)
	if err != nil {
		logs.Info(err.Error())
		return nil, errors.New("Icon is not found")

	}
	_, err = io.Copy(newFile, rc)
	if err == nil {
		p.AppIcon = appiconpath
	}
	defer newFile.Close()
	return p, nil

}

// CreateUdidPlist  生成获取udid的描述文件
func (a *AppleService) CreateUdidPlist(cfBundleDisplayName string, path string, hostudidurl string) (signedpath string, err error) {

	f, _ := os.OpenFile("mobileconfig/unsigned.mobileconfig", os.O_RDONLY, 0600)
	defer f.Close()
	contentByte, _ := ioutil.ReadAll(f)
	str := utils.Bytes2str(contentByte)
	str2 := strings.Replace(str, "udidurl", hostudidurl, -1)
	str3 := strings.Replace(str2, "显示名称", cfBundleDisplayName, -1)
	unsignedpath := utils.SplicingString(filepath.Dir(path), "/unsigned.mobileconfig")
	signedpath = utils.SplicingString(filepath.Dir(path), "/signed.mobileconfig")
	logs.Info("描述文件保存:", signedpath)
	f2, _ := os.Create(unsignedpath)
	defer f2.Close()
	f2.Write(utils.Str2bytes(str3))
	//使用显示名称生成描述文件获取udid 到对应app目录下 名为signed.mobileconfig
	// openssl smime -sign -in unsigned.mobileconfig -out signed.mobileconfig -signer server.crt -inkey private.key -certfile root.crt -outform der -nodetach
	logs.Info("openssl", "smime", "-sign", "-in", unsignedpath, "-out", signedpath, "-signer", "mobileconfig/server.crt", "-inkey", "mobileconfig/private.key", "-certfile", "mobileconfig/root.crt", "-outform", "der", "-nodetach")
	_, err = exec.Command("openssl", "smime", "-sign", "-in", unsignedpath, "-out", signedpath, "-signer", "mobileconfig/server.crt", "-inkey", "mobileconfig/private.key", "-certfile", "mobileconfig/root.crt", "-outform", "der", "-nodetach").Output()
	if err != nil {
		logs.Info("执行命令错误", err.Error())
		return "", err
	}
	os.Remove(unsignedpath)
	return signedpath, nil
}

// RegistDevice 注册一个新设备
func (a *AppleService) RegistDevice(appleaccount *models.AppleAccount, deviceName string, udid string) (appleID string, err error) {
	signedToken, err := utils.GetAppleToken(appleaccount.IssuerID, appleaccount.KeyID, appleaccount.P8filepath)
	if err != nil {
		return "", errors.New("生成token错误")
	}
	data := RegistDevice{}
	data.Data.Type = "devices"
	data.Data.Attributes.Name = deviceName
	data.Data.Attributes.UDID = udid
	data.Data.Attributes.Platform = "IOS"
	datajson, _ := jsoniter.Marshal(data)
	req := httplib.Post(registdeviceURL)
	authorization := utils.SplicingString("Bearer ", signedToken)
	req.Header("Authorization", authorization)
	req.Header("Content-Type", "application/json")
	req.Body(utils.Bytes2str(datajson))
	res, err := req.Response()
	if err != nil {
		return "", errors.New("苹果接口访问错误")
	}
	if res.StatusCode == 400 || res.StatusCode == 403 || res.StatusCode == 409 {
		appleerror := AppleAPIError{}
		req.ToJSON(&appleerror)
		logs.Info(appleerror)
		return "", errors.New("接口错误:" + appleerror.Errors[0].Detail)
	} else if res.StatusCode == 201 {
		bytes, _ := req.Bytes()
		appleID = jsoniter.Get(bytes, "data").Get("id").ToString()
		//添加设备成功
		return appleID, nil
	} else if res.StatusCode == 401 {
		//帐号异常
		a.AppleAccountOdd(appleaccount)
		return "", errors.New("帐号异常")
	} else {
		return "", errors.New("苹果接口访问错误")
	}
}

// RegistBundleid 注册一个新bundleid
func (a *AppleService) RegistBundleid(appleaccount *models.AppleAccount, bundleid string) (appleID string, err error) {
	signedToken, err := utils.GetAppleToken(appleaccount.IssuerID, appleaccount.KeyID, appleaccount.P8filepath)
	if err != nil {
		return "", errors.New("生成token错误")
	}
	data := RegistBundleid{}
	data.Data.Type = "bundleIds"
	data.Data.Attributes.Name = strings.Replace(bundleid, ".", "", -1)
	data.Data.Attributes.Identifier = bundleid
	data.Data.Attributes.Platform = "IOS"
	datajson, _ := jsoniter.Marshal(data)
	req := httplib.Post(registBundleidURL)
	authorization := utils.SplicingString("Bearer ", signedToken)
	req.Header("Authorization", authorization)
	req.Header("Content-Type", "application/json")
	req.Body(utils.Bytes2str(datajson))
	res, err := req.Response()
	if err != nil {
		return "", errors.New("苹果接口访问错误")
	}
	if res.StatusCode == 400 || res.StatusCode == 403 || res.StatusCode == 409 {
		appleerror := AppleAPIError{}
		req.ToJSON(&appleerror)
		logs.Info(appleerror)
		return "", errors.New("接口错误:" + appleerror.Errors[0].Detail)
	} else if res.StatusCode == 201 {
		bytes, _ := req.Bytes()
		appleID = jsoniter.Get(bytes, "data").Get("id").ToString()
		logs.Info(appleID)
		//注册bundleid成功
		return appleID, nil
	} else if res.StatusCode == 401 {
		//帐号异常
		a.AppleAccountOdd(appleaccount)
		return "", errors.New("帐号异常")
	} else {
		return "", errors.New("苹果接口访问错误")
	}
}

// RegistProfile 注册描述文件
// appleDeviceID 苹果系统中的设备ID
// appleBundleID 苹果系统中的bundleid
// appleCerID 苹果系统中的证书ID
func (a *AppleService) RegistProfile(appleaccount *models.AppleAccount, appleDeviceID string, appleBundleID string, appleCerID string) (registProfileResponse RegistProfileResponse, err error) {
	signedToken, err := utils.GetAppleToken(appleaccount.IssuerID, appleaccount.KeyID, appleaccount.P8filepath)
	if err != nil {
		return registProfileResponse, errors.New("生成token错误")
	}
	data := RegistProfile{}
	data.Data.Type = "profiles"
	data.Data.Attributes.Name = utils.GetRandString(6)
	data.Data.Attributes.ProfileType = "IOS_APP_ADHOC"
	data.Data.Relationships.BundleID.Data.ID = appleBundleID
	data.Data.Relationships.BundleID.Data.Type = "bundleIds"
	data.Data.Relationships.Certificates.Data = append(data.Data.Relationships.Certificates.Data, Edata{appleCerID, "certificates"})
	data.Data.Relationships.Devices.Data = append(data.Data.Relationships.Devices.Data, Edata{appleDeviceID, "devices"})
	datajson, _ := jsoniter.Marshal(data)
	req := httplib.Post(registProfileURL)
	authorization := utils.SplicingString("Bearer ", signedToken)
	req.Header("Authorization", authorization)
	req.Header("Content-Type", "application/json")
	req.Body(utils.Bytes2str(datajson))
	res, err := req.Response()
	if err != nil {
		return registProfileResponse, errors.New("苹果接口访问错误")
	}
	if res.StatusCode == 400 || res.StatusCode == 403 || res.StatusCode == 409 {
		appleerror := AppleAPIError{}
		req.ToJSON(&appleerror)
		logs.Info(appleerror)
		return registProfileResponse, errors.New("接口错误:" + appleerror.Errors[0].Detail)
	} else if res.StatusCode == 201 {
		req.ToJSON(&registProfileResponse)
		return registProfileResponse, nil
	} else if res.StatusCode == 401 {
		//帐号异常
		a.AppleAccountOdd(appleaccount)
		return registProfileResponse, errors.New("帐号异常")
	} else {
		return registProfileResponse, errors.New("苹果接口访问错误")
	}
}

// AppleAccountOdd 帐号异常
func (a *AppleService) AppleAccountOdd(appleaccount *models.AppleAccount) {
	logs.Info("apple 帐号异常：", appleaccount)
	//更新帐号状态
	appleaccount.Status = 0
	appleaccount.UpdateAccount(appleaccount)
	signeds, _ := models.NewSignedInfo().GetSignedInfoByApple(appleaccount.ID)
	//删除已经签名的文件
	files := []string{}
	for _, v := range signeds {
		p := filepath.Dir(v.SignedipaPath)
		files = append(files, v.SignedipaPath)
		go os.RemoveAll(p)
	}
	go ossservice.DeleteFiles(files)
	//更新签名记录
	models.NewSignedInfo().SetOddSigned(appleaccount.ID)
}
