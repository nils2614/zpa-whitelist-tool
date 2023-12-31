package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Struct to hold information from Zscaler api
type response struct {
	CloudName string `json:"Cloud Name"`
	Content   []struct {
		IPProtocol string   `json:"IP Protocol"`
		Port       int      `json:"Port"`
		Source     string   `json:"Source"`
		Domains    string   `json:"Domains"`
		IPs        []string `json:"IPs"`
		DateAdded  string   `json:"Date Added"`
	} `json:"content"`
}

func main() {
	var (
		myurl     string = "https://api.config.zscaler.com/zscaler.net/zpa/json" // url of the Zscaler api
		rgNameTf  string = "whitelist-dev"                                       // name of the resource group in Terraform
		nsgNameTf string = "whitelist-dev"                                       // name of the network security group in Terraform
		nsgNameAz string = "zpa-whitelist-dev"                                   // name of the network security group in Azure
	)

	// Request data and store body in var.body
	resp, err := http.Get(myurl)
	if err != nil {
		fmt.Println("No response from request")
	}

	// Close ReadAll of resp.Body
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body of data request")
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	// Unmarshal JSON from var.body to struct.result
	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	//printResponse(result) // Debug function
	generateTerraform(rgNameTf, nsgNameTf, nsgNameAz, result)
}

func printResponse(r response) {
	fmt.Println("")
	for i := 0; i < len(r.Content); i++ {
		fmt.Print(i)
		fmt.Print(": " + r.Content[i].IPProtocol + ", Port ")
		fmt.Print(r.Content[i].Port)
		fmt.Print(", added on " + r.Content[i].DateAdded + "\n")

		fmt.Println(r.Content[i].IPs)
		fmt.Println("")
	}
}

func writeToFile(lines []string) {
	file, err := os.Create("output.tf")
	if err != nil {
		fmt.Println("Error creating file on disk")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file")
		}
	}(file)

	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			fmt.Println("Error writing to file on disk")
		}
	}

}

func generateTerraform(rgNameTf string, nsgNameTf string, nsgNameAz string, r response) {
	var tfCode []string // empty slice to store Terraform code line by line

	// Terraform nsg definition code
	tfCode = append(tfCode, "# Terraform code generated by zscaler-address-tool")
	tfCode = append(tfCode, "resource \"azurerm_network_security_group\" \""+nsgNameTf+"\" {")
	tfCode = append(tfCode, "  name                = \""+nsgNameAz+"\"")
	tfCode = append(tfCode, "  location            = azurerm_resource_group."+rgNameTf+".location")
	tfCode = append(tfCode, "  resource_group_name = azurerm_resource_group."+rgNameTf+".name")
	tfCode = append(tfCode, "")

	// DenyAllOutBound security rule
	tfCode = append(tfCode, generateSecurityRule("DenyAllOutBound", 4000, "Outbound", "Deny", "*", "*", "*")...)

	// Add whitelist security rules from Zscaler api
	tfCode = append(tfCode, appendWhitelistRules(r, 2001)...)

	// Close Terraform nsg definition
	tfCode = append(tfCode, "}")
	writeToFile(tfCode)
	fmt.Println("Terraform code successfully generated!")
}

func appendWhitelistRules(r response, priority int) []string {
	var whitelistRules []string

	// Iterate over all IPs and append security rules for them (TCP and UDP)
	for i := 0; i < len(r.Content); i++ {
		fmt.Println("Rules are being generated for IP Block " + strconv.Itoa(i+1) + " (" + r.Content[i].DateAdded + ").")
		for j := 0; j < len(r.Content[i].IPs); j++ {
			var ruleName string = "AllowZscaler" + "-" + strconv.Itoa(i+1) + "-" + strconv.Itoa(j+1)
			whitelistRules = append(whitelistRules, generateSecurityRule(ruleName, priority, "Outbound", "Allow", "*", "443", r.Content[i].IPs[j])...)
			priority++
		}
	}

	return whitelistRules
}

func generateSecurityRule(name string, priority int, direction string, access string, protocol string, port string, ip string) []string {
	var securityRule []string
	securityRule = append(securityRule, "  security_rule {")
	securityRule = append(securityRule, "    name                       = \""+name+"\"")
	securityRule = append(securityRule, "    priority                   = "+strconv.Itoa(priority))
	securityRule = append(securityRule, "    direction                  = \""+direction+"\"")
	securityRule = append(securityRule, "    access                     = \""+access+"\"")
	securityRule = append(securityRule, "    protocol                   = \""+protocol+"\"")
	securityRule = append(securityRule, "    source_port_range          = \"*\"")
	securityRule = append(securityRule, "    destination_port_range     = \""+port+"\"")
	securityRule = append(securityRule, "    source_address_prefix      = \"*\"")
	securityRule = append(securityRule, "    destination_address_prefix = \""+ip+"\"")
	securityRule = append(securityRule, "  }")
	return securityRule
}
