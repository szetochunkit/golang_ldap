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
	AddReq.Attribute("userPrincipalName", []string{fmt.Sprintf("%s@lexintest.com", Username)})
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

func AddGroupMember(l *ldap.Conn, Username string, Groupname string) string {
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

func RemoveGroupMember(l *ldap.Conn, Username string, Groupname string) string {
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
