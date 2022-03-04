package models

import "github.com/beego/beego/v2/client/orm"

// AppleCertificates 苹果cer文件
type AppleCertificates struct {
	ID                 int    `json:"id" orm:"column(id);auto"`
	AccountID          int    `json:"account_id" orm:"column(account_id);description(苹果帐号ID)"`
	AppleID            string `json:"appleid" orm:"column(appleid);size(255);description(苹果那边的id)"`
	SerialNumber       string `json:"serialNumber" orm:"column(serial_number);size(255);description(证书序列号)"`
	CertificateContent string `json:"certificateContent" orm:"column(certificate_content);type(text);description(cer文件内容)"`
	CertificatePath    string `json:"certificatePath" orm:"column(certificate_path);type(255);description(cer文件path)"`
	Name               string `json:"name" orm:"column(display_name);size(255);description(证书名字)"`
	CsrContent         string `json:"csrContent" orm:"column(csr_content);type(text);description(csr文件内容)"`
	Platform           string `json:"platform"  orm:"column(platform);size(255);description(证书可用设备类型)"`
	ExpirationDate     string `json:"expirationDate" orm:"column(expiration_date);size(255);description(过期时间)"`
	CertificateType    string `json:"certificateType" orm:"column(certificate_type);size(255);description(证书类型)"`
}

// TableName ..
func (b *AppleCertificates) TableName() string {
	return "apple_certificates"
}

func init() {
	orm.RegisterModel(new(AppleCertificates))
}

// NewAppleCertificates ...
func NewAppleCertificates() *AppleCertificates {
	return &AppleCertificates{}
}

// InsertCertificates 新增一批证书文件
func (b *AppleCertificates) InsertCertificates(certificates []AppleCertificates) {
	o := orm.NewOrm()
	o.InsertMulti(len(certificates), &certificates)
}
