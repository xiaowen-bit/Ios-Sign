package appleservice

import "kksigncustom/models"

// AppleAPIError 苹果接口错误
type AppleAPIError struct {
	Errors []struct {
		Status string `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

// RegistDevice 注册一个新设备
type RegistDevice struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Name     string `json:"name"`
			Platform string `json:"platform"`
			UDID     string `json:"udid"`
		} `json:"attributes"`
	} `json:"data"`
}

// RegistBundleid 注册一个新设备
type RegistBundleid struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			//显示的名字
			Name string `json:"name"`
			//固定值IOS
			Platform string `json:"platform"`
			// bundelid
			Identifier string `json:"identifier"`
		} `json:"attributes"`
	} `json:"data"`
}

// AppleAPIDevice apple 返回的数据结构
type AppleAPIDevice struct {
	Data []struct {
		Type        string             `json:"type"`
		ID          string             `json:"id"`
		AppleDevice models.AppleDevice `json:"attributes"`
		Links       struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	Meta struct {
		Paging struct {
			Total int `json:"total"`
			Limit int `json:"limit"`
		} `json:"paging"`
	} `json:"meta"`
}

// Edata 最下一成数据
type Edata struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// RegistProfile 注册一个新描述文件
type RegistProfile struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			//显示的名字
			Name string `json:"name"`
			//固定值IOS
			ProfileType string `json:"profileType"`
		} `json:"attributes"`
		Relationships struct {
			BundleID struct {
				Data struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				} `json:"data"`
			} `json:"bundleId"`
			Certificates struct {
				Data []Edata `json:"data"`
			} `json:"certificates"`
			Devices struct {
				Data []Edata `json:"data"`
			} `json:"devices"`
		} `json:"relationships"`
	} `json:"data"`
}

// RegistProfileResponse 注册新描述文件的返回结果
type RegistProfileResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			ProfileState   string `json:"profileState"`
			CreatedDate    string `json:"createdDate"`
			ProfileType    string `json:"profileType"`
			Name           string `json:"name"`
			ProfileContent string `json:"profileContent"`
			UUID           string `json:"uuid"`
			Platform       string `json:"platform"`
			ExpirationDate string `json:"expirationDate"`
		} `json:"attributes"`
		Relationships struct {
			BundleID struct {
				Data  Edata `json:"data"`
				Links struct {
					Self    string `json:"self"`
					Related string `json:"related"`
				} `json:"links"`
			} `json:"bundleId"`
			Certificates struct {
				Meta struct {
					Paging struct {
						Total int   `json:"total"`
						Limit int64 `json:"limit"`
					} `json:"paging"`
				} `json:"meta"`
				Data  []Edata `json:"data"`
				Links struct {
					Self    string `json:"self"`
					Related string `json:"related"`
				} `json:"links"`
			} `json:"certificates"`
			Devices struct {
				Meta struct {
					Paging struct {
						Total int   `json:"total"`
						Limit int64 `json:"limit"`
					} `json:"paging"`
				} `json:"meta"`
				Data  []Edata `json:"data"`
				Links struct {
					Self    string `json:"self"`
					Related string `json:"related"`
				} `json:"links"`
			} `json:"devices"`
		} `json:"relationships"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}
