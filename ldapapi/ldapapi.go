package ldapapi

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
	"io/ioutil"
)

const (
	CertPath = "cert/certnew.crt"
	LdapHost = "1.1.1.1"
	LdapPort = "636"
	BindUser = "CN=golang,CN=Users,DC=test,DC=com"
	BindPWD  = "123"
	BaseDN   = "DC=test,DC=com"
)

func filetimeToFormattedString(filetime uint64) string {
	// 1601-01-01 00:00:00 UTC 对应的 FileTime 值（以纳秒为单位）
	const filetimeEpoch = 116444736000000000

	// 将 FileTime 转换为纳秒并减去起始的差值
	ns := int64(filetime) - filetimeEpoch

	// 将纳秒转换为秒
	secs := ns / 10000000 // 100纳秒 = 0.0000001秒，所以10^7纳秒 = 1秒

	// 使用秒数创建 time.Time 对象
	timeObj := time.Unix(secs, 0)

	// 将 time.Time 对象格式化为所需的字符串格式
	return timeObj.Format("2006/1/2 15:04:05")
}
func formattedStringToLDAPTimestamp(formattedTime string) string {
	// 解析格式化的日期时间字符串为 time.Time 对象
	parsedTime, err := time.Parse("2006-1-2 15:04:05", formattedTime)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return ""
	}

	// 1601-01-01 00:00:00 UTC 对应的纳秒数
	const filetimeEpoch = 116444736000000000

	// 获取时间对象对应的纳秒数并计算与FileTime的起始值的差
	totalNanoseconds := parsedTime.UnixNano()
	ldapTimestamp := totalNanoseconds/100 + filetimeEpoch

	// 转换为字符串并返回18-digit LDAP timestamp格式
	return fmt.Sprintf("%018d", ldapTimestamp)
}

func BindLdap() (*ldap.Conn, error) {
	CaCert, err := ioutil.ReadFile(CertPath)
	if err != nil {
		fmt.Println("read cert error", err.Error())
		return nil, err
	}
	fmt.Println("read cert OK")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(CaCert)
	tlsConfig := &tls.Config{
		ClientCAs:          caCertPool,
		InsecureSkipVerify: true,
	}
	//LdapAddr := LdapHost + ":" + LdapPort
	LdapAddr := fmt.Sprintf("%s:%s", LdapHost, LdapPort)
	l, err := ldap.DialTLS("tcp", LdapAddr, tlsConfig)
	if err != nil {
		fmt.Println("connect ladp error:", err.Error())
		return nil, err
	}

	fmt.Println("connect OK")
	_, err = l.SimpleBind(&ldap.SimpleBindRequest{
		Username: BindUser,
		Password: BindPWD,
	})
	if err != nil {
		fmt.Println("bind user error:", err.Error())
		l.Close()
		return nil, err
	}
	fmt.Println("bind OK")
	//fmt.Println("l:", reflect.TypeOf(l))
	return l, nil

}

func GetUserInfo(l *ldap.Conn, Username string) (*ldap.Entry, string) {
	searchRequest := ldap.NewSearchRequest(
		BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", username),
		fmt.Sprintf("(&(objectclass=user)(sAMAccountName=%s))", Username),
		//[]string{"dn"},
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		//fmt.Println("search user error:", err.Error())
		return nil, "search user error"
	}
	if len(sr.Entries) == 0 {
		//fmt.Println("User does not exist")
		return nil, "User does not exist"
	}
	if len(sr.Entries) > 1 {
		//fmt.Println("too many entries returned")
		return nil, "too many entries returned"
	}
	UserInfo := sr.Entries[0]
	//fmt.Println("UserDn:", reflect.TypeOf(UserDn))
	return UserInfo, ""
}

func ModifyUser(l *ldap.Conn, Username string, Attribute string, Value string) string {
	UserInfo, ReturnStr := GetUserInfo(l, Username)
	if ReturnStr != "" {
		fmt.Println(ReturnStr)
		return ReturnStr
	}
	UserDN := UserInfo.DN
	ModifyReq := ldap.NewModifyRequest(UserDN, nil)
	ModifyReq.Replace(Attribute, []string{Value})
	//err = l.Modify(UserModifyReq)
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func SetNewPassword(l *ldap.Conn, Username string, NewPassword string) string {
	UserInfo, ReturnStr := GetUserInfo(l, Username)
	if ReturnStr != "" {
		fmt.Println(ReturnStr)
		return ReturnStr
	}
	UserDN := UserInfo.DN
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	PwdEncoded, err := utf16.NewEncoder().String("\"" + NewPassword + "\"")
	if err != nil {
		fmt.Println("password unicode error:", err.Error())
		return "password unicode error"
	}
	ModifyReq := ldap.NewModifyRequest(UserDN, nil)
	ModifyReq.Replace("unicodePwd", []string{PwdEncoded})
	err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("reset password error:", err.Error())
		return "reset password error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func CreatUser(l *ldap.Conn, Username string, OuPath string) string {
	UserDN := fmt.Sprintf("cn=%s,%s", Username, OuPath)
	AddReq := ldap.NewAddRequest(UserDN, []ldap.Control{})
	AddReq.Attribute("objectClass", []string{"top", "organizationalPerson", "user", "person"})
	AddReq.Attribute("sAMAccountName", []string{Username})
	AddReq.Attribute("userPrincipalName", []string{fmt.Sprintf("%s@test.com", Username)})
	var err = l.Add(AddReq)
	if err != nil {
		fmt.Println("error adding service:", AddReq, err)
		return "creat error"
	}
	fmt.Println("creat OK")
	return "creat ok"
}

func MoveUserToOU(l *ldap.Conn, Username string, OuPath string) string {
	UserInfo, ReturnStr := GetUserInfo(l, Username)
	if ReturnStr != "" {
		fmt.Println(ReturnStr)
		return ReturnStr
	}
	UserDN := UserInfo.DN
	CN := fmt.Sprintf("CN=%s", Username)
	MoveReq := ldap.NewModifyDNRequest(UserDN, CN, true, OuPath)
	var err = l.ModifyDN(MoveReq)
	if err != nil {
		fmt.Println("move user error", err.Error())
		return "move user error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func GetGroupInfo(l *ldap.Conn, Groupname string) (*ldap.Entry, string) {
	searchRequest := ldap.NewSearchRequest(
		BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", username),
		fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", Groupname),
		//[]string{"dn"},
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		//fmt.Println("search user error:", err.Error())
		return nil, "search user error"
	}
	if len(sr.Entries) == 0 {
		//fmt.Println("User does not exist")
		return nil, "User does not exist"
	}
	if len(sr.Entries) > 1 {
		//fmt.Println("too many entries returned")
		return nil, "too many entries returned"
	}
	GroupInfo := sr.Entries[0]
	//fmt.Println("UserDn:", reflect.TypeOf(UserDn))
	return GroupInfo, ""
}

func AddUserTOGroup(l *ldap.Conn, Username string, Groupname string) string {
	UserInfo, UserReturnINfo := GetUserInfo(l, Username)
	if UserReturnINfo != "" {
		return UserReturnINfo
	}
	UserDn := UserInfo.DN
	GroupInfo, GroupReturnINfo := GetGroupInfo(l, Groupname)
	if GroupReturnINfo != "" {
		return GroupReturnINfo
	}
	GroupDn := GroupInfo.DN
	ModifyReq := ldap.NewModifyRequest(GroupDn, nil)
	ModifyReq.Add("member", []string{UserDn})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func RemoveUserFromGroup(l *ldap.Conn, Username string, Groupname string) string {
	UserInfo, UserReturnINfo := GetUserInfo(l, Username)
	if UserReturnINfo != "" {
		return UserReturnINfo
	}
	UserDn := UserInfo.DN
	GroupInfo, GroupReturnINfo := GetGroupInfo(l, Groupname)
	if GroupReturnINfo != "" {
		return GroupReturnINfo
	}
	GroupDn := GroupInfo.DN
	ModifyReq := ldap.NewModifyRequest(GroupDn, nil)
	ModifyReq.Delete("member", []string{UserDn})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func ModifyGroup(l *ldap.Conn, Groupname string, Attribute string, Value string) string {
	GroupInfo, GroupReturnINfo := GetGroupInfo(l, Groupname)
	if GroupReturnINfo != "" {
		return GroupReturnINfo
	}
	GroupDn := GroupInfo.DN
	ModifyReq := ldap.NewModifyRequest(GroupDn, nil)
	ModifyReq.Replace(Attribute, []string{Value})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func CreatGroup(l *ldap.Conn, Groupname string, OuPath string) string {
	GroupDN := fmt.Sprintf("cn=%s,%s", Groupname, OuPath)
	AddReq := ldap.NewAddRequest(GroupDN, []ldap.Control{})
	AddReq.Attribute("objectClass", []string{"top", "group"})
	AddReq.Attribute("sAMAccountName", []string{Groupname})
	AddReq.Attribute("displayname", []string{Groupname})
	var err = l.Add(AddReq)
	if err != nil {
		fmt.Println("error adding service:", AddReq, err)
		return "creat error"
	}
	fmt.Println("creat OK")
	return "creat ok"
}

func GetOrganizationalUnitInfo(l *ldap.Conn, OuDN string) (*ldap.Entry, string) {
	searchRequest := ldap.NewSearchRequest(
		BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectclass=organizationalUnit)(DistinguishedName=%s))", OuDN),
		//[]string{"dn"},
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		//fmt.Println("search user error:", err.Error())
		return nil, "search user error"
	}
	if len(sr.Entries) == 0 {
		//fmt.Println("User does not exist")
		return nil, "User does not exist"
	}
	if len(sr.Entries) > 1 {
		//fmt.Println("too many entries returned")
		return nil, "too many entries returned"
	}
	OuInfo := sr.Entries[0]
	//fmt.Println("UserDn:", reflect.TypeOf(UserDn))
	return OuInfo, ""
}

func CreatOrganizationalUnit(l *ldap.Conn, OuName string, OuPath string) string {
	OuDN := fmt.Sprintf("OU=%s,%s", OuName, OuPath)
	AddReq := ldap.NewAddRequest(OuDN, []ldap.Control{})
	AddReq.Attribute("objectClass", []string{"top", "organizationalUnit"})
	AddReq.Attribute("name", []string{OuName})
	var err = l.Add(AddReq)
	if err != nil {
		fmt.Println("error adding service:", AddReq, err)
		return "creat error"
	}
	fmt.Println("creat OK")
	return "creat ok"
}

func ModifyOrganizationalUnit(l *ldap.Conn, OuDN string, Attribute string, Value string) string {
	ModifyReq := ldap.NewModifyRequest(OuDN, nil)
	ModifyReq.Replace(Attribute, []string{Value})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func GetOuAllUsers(l *ldap.Conn, OuPath string) ([]*ldap.Entry, string) {
	searchRequest := ldap.NewSearchRequest(
		OuPath,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectclass=user))"),
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, "search user error"
	}
	if len(sr.Entries) == 0 {
		//fmt.Println("User does not exist")
		return nil, "User does not exist"
	}

	return sr.Entries, ""
}

func GetOuAllGroups(l *ldap.Conn, OuPath string) ([]*ldap.Entry, string) {
	searchRequest := ldap.NewSearchRequest(
		OuPath,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectclass=group))"),
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, "search group error"
	}
	if len(sr.Entries) == 0 {
		//fmt.Println("User does not exist")
		return nil, "group does not exist"
	}
	return sr.Entries, ""
}

func AddSubGroupToGroup(l *ldap.Conn, SubGroupname string, Groupname string) string {
	SubGroupInfo, SubGroupReturnINfo := GetGroupInfo(l, SubGroupname)
	if SubGroupReturnINfo != "" {
		return SubGroupReturnINfo
	}
	GroupInfo, GroupReturnINfo := GetGroupInfo(l, Groupname)
	if GroupReturnINfo != "" {
		return GroupReturnINfo
	}
	SubGroupDn := SubGroupInfo.DN
	GroupDn := GroupInfo.DN
	ModifyReq := ldap.NewModifyRequest(GroupDn, nil)
	ModifyReq.Add("member", []string{SubGroupDn})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func RemoveSubGroupFromGroup(l *ldap.Conn, SubGroupname string, Groupname string) string {
	SubGroupInfo, SubGroupReturnINfo := GetGroupInfo(l, SubGroupname)
	if SubGroupReturnINfo != "" {
		return SubGroupReturnINfo
	}
	GroupInfo, GroupReturnINfo := GetGroupInfo(l, Groupname)
	if GroupReturnINfo != "" {
		return GroupReturnINfo
	}
	SubGroupDn := SubGroupInfo.DN
	GroupDn := GroupInfo.DN
	ModifyReq := ldap.NewModifyRequest(GroupDn, nil)
	ModifyReq.Delete("member", []string{SubGroupDn})
	var err = l.Modify(ModifyReq)
	if err != nil {
		fmt.Println("modify error", err.Error())
		return "modify error"
	}
	fmt.Println("modify OK")
	return "modify OK"
}

func SetPasswordNeverExpires(l *ldap.Conn, Username string) string {
	//ModifyUser(l *ldap.Conn, Username string, Attribute string, Value string)
	Attribute := "userAccountControl"
	Value := "66048"
	ReturnStr := ModifyUser(l, Username, Attribute, Value)
	return ReturnStr
}

func CancelPasswordNeverExpires(l *ldap.Conn, Username string) string {
	//ModifyUser(l *ldap.Conn, Username string, Attribute string, Value string)
	Attribute := "userAccountControl"
	Value := "512"
	ReturnStr := ModifyUser(l, Username, Attribute, Value)
	return ReturnStr
}

func EnableUser(l *ldap.Conn, Username string) string {
	Attribute := "userAccountControl"
	Value := "512"
	ReturnStr := ModifyUser(l, Username, Attribute, Value)
	return ReturnStr
}

func DisableUser(l *ldap.Conn, Username string) string {
	Attribute := "userAccountControl"
	Value := "514"
	ReturnStr := ModifyUser(l, Username, Attribute, Value)
	return ReturnStr
}

func SetUserManager(l *ldap.Conn, Username string, ManagerUsername string) string {
	Attribute := "manager"
	ManagerUserInfo, ManagerUserReturnStr := GetUserInfo(l, ManagerUsername)
	if ManagerUserReturnStr != "" {
		fmt.Println(ManagerUserReturnStr)
		return ManagerUserReturnStr
	}
	ManagerUserDN := ManagerUserInfo.DN
	ReturnStr := ModifyUser(l, Username, Attribute, ManagerUserDN)
	return ReturnStr
}

func SetAccountExpirationDate(l *ldap.Conn, Username string, ExpirationDate string) string {
	ldapTimestamp := formattedStringToLDAPTimestamp(ExpirationDate)
	// 输出转换后的18-digit LDAP timestamp值
	fmt.Println("18-digit LDAP Timestamp:", ldapTimestamp)
	Attribute := "accountExpires"
	ReturnStr := ModifyUser(l, Username, Attribute, ldapTimestamp)
	return ReturnStr
}

func SetAccountNotExpired(l *ldap.Conn, Username string) string {
	Attribute := "accountExpires"
	ReturnStr := ModifyUser(l, Username, Attribute, "0")
	return ReturnStr
}
