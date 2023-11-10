package main

import (
	"fmt"
	"ldap_golang/ldapapi"
)

func main() {
	l, err := ldapapi.BindLdap()
	if err != nil {
		fmt.Println("bind user error:", err.Error())
		//return "bind user error"
	}
	//Username := "ssbb"

	//get user info
	/*
		UserInfo, ReturnStr := ldapapi.GetUserInfo(l, Username)
		if ReturnStr != "" {
			fmt.Println(ReturnStr)
			//return ReturnStr
		}
		UserInfo.Print()
		//userDepartment := UserInfo.GetAttributeValue("department")
		//fmt.Println(userDepartment)
	*/

	//add user
	/*
		Username := "asdfg"
		OuPath := ""
		ldapapi.AddUser(l, Username, OuPath)
	*/

	//set user password
	/*
		NewPassword := ""
		ldapapi.SetNewPassword(l, Username, NewPassword)
	*/

	//modify user
	/*
		Attribute := "description"
		Value := "asdf123"
		ldapapi.ModifyUser(l, Username, Attribute, Value)
	*/

	//enable user
	/*
		Attribute := "userAccountControl"
		Value := "512"
		ldapapi.ModifyUser(l, Username, Attribute, Value)
	*/

	//disable user
	/*
		Attribute := "userAccountControl"
		Value := "514"
		ldapapi.ModifyUserAttribute(Username, Attribute, Value)
	*/

	//move user to OU
	/*
		OuPath := ""
		ldapapi.MoveUserToOU(l, Username, OuPath)
	*/

	//get group info
	/*
		Groupname := ""
		GroupInfo, ReturnINfo := ldapapi.GetGroupInfo(l, Groupname)
		if ReturnINfo != "" {
			fmt.Println(ReturnINfo)
		}
		//GroupInfo.Print()
		member := GroupInfo.GetAttributeValues("member")
		fmt.Println(member)
	*/

}
