# Zscaler ZPA Whitelist Tool
Simple tool to generate a Terraform definition of an Azure nsg to only allow outgoing traffic to the specified Zscaler ZPA Networks.

### Variables in func main
myurl: This is the URL of the Zscaler API  
rgNameTf: The name of the resource group where the NSG should reside used for reference in Terraform  
nsgNameTf: The name of the NSG used for reference in Terraform  
nsgNameAz: The name of the NSG to be used in Azure  

## to do
- Read variables from config file instead of defining them in the main function
- Let user set starting priorities for security values
