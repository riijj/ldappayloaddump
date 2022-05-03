package main

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%s <url> [out_file]\nFor example : %s ldap://127.0.0.1:1389/Basic/Command/whoami\n", os.Args[0], os.Args[0])
		return
	}

	targetUrl := os.Args[1]
	URL ,err := url.Parse(targetUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	ldapDialUrl := "ldap://" + URL.Host
	searchPath := URL.Path
	if len(searchPath) == 0 {
		searchPath = "/"
	}
	var outFile string = "data.class"
	if len(os.Args) >= 3 {
		outFile = os.Args[2]
	}


	conn, err := ldap.DialURL(ldapDialUrl)
	if err != nil {
		fmt.Printf("ldap.DialUrl(%s) is failed : %s\n", ldapDialUrl, err)
		return
	}
	bindReq := &ldap.SimpleBindRequest{
		Username:           "",
		Password:           "",
		AllowEmptyPassword: true,
	}

	conn.SimpleBind(bindReq)

	if len(searchPath) > 1 && searchPath[0] == '/' {
		searchPath = searchPath[1:]
	}
	newSearch := ldap.NewSearchRequest(searchPath, ldap.ScopeBaseObject, ldap.DerefAlways, 0, 0, false, "(objectClass=*)", []string{}, nil)
	searchResult, err := conn.Search(newSearch)
	if err != nil {
		fmt.Printf("[-] ldap search failed : %s\n", err)
		return
	}

	var data []byte
	for _, entry := range searchResult.Entries {

		codeBaseValue := entry.GetAttributeValue("javaCodeBase")
		factoryValue := entry.GetAttributeValue("javaFactory")

		if len(codeBaseValue) > 0 && len(factoryValue) > 0 {
			reqPath := strings.ReplaceAll(factoryValue,".","/")
			classReqUrl := fmt.Sprintf("%s%s.class",codeBaseValue,reqPath)
			fmt.Printf("[+] Dump class from %s\n",classReqUrl)
			resp,err := http.Get(classReqUrl)
			if err != nil {
				fmt.Printf("[-] http.Get() is failed :%s\n",err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				fmt.Printf("[-] http.Get() : HTTP %d \n",resp.StatusCode)
			} else {
				content,_ := ioutil.ReadAll(resp.Body)
				data = content
			}
		} else {
			for _, attr := range entry.Attributes {
				if attr.Name == "javaSerializedData" {
					fmt.Printf("[+] Found serialized data, dump....\n")
					data = attr.ByteValues[0]
				}
			}
		}
	}



	ioutil.WriteFile(outFile,data,0777)
	fmt.Printf("[+] Dumped data :%s\n",outFile)
	fmt.Printf("[+] All done!\n")
}
