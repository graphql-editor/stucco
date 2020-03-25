package deployment

// FunctionLocations list valid function location deployments for stucco
var (
	FunctionLocations = []struct {
		Name, Slug string
	}{
		{Name: "East Asia", Slug: "eastasia"},
		{Name: "Southeast Asia", Slug: "southeastasia"},
		{Name: "Australia East", Slug: "australiaeast"},
		{Name: "Australia Southeast", Slug: "australiasoutheast"},
		{Name: "Brazil South", Slug: "brazilsouth"},
		{Name: "Canada Central", Slug: "canadacentral"},
		{Name: "North Europe", Slug: "northeurope"},
		{Name: "West Europe", Slug: "westeurope"},
		{Name: "France Central", Slug: "francecentral"},
		{Name: "West India", Slug: "westindia"},
		{Name: "Japan East", Slug: "japaneast"},
		{Name: "Japan West", Slug: "japanwest"},
		{Name: "Korea Central", Slug: "koreacentral"},
		{Name: "Korea South", Slug: "koreasouth"},
		{Name: "Norway East", Slug: "norwayeast"},
		{Name: "UK South", Slug: "uksouth"},
		{Name: "UK West", Slug: "ukwest"},
		{Name: "Central US", Slug: "centralus"},
		{Name: "East US", Slug: "eastus"},
		{Name: "East US 2", Slug: "eastus2"},
		{Name: "North Central US", Slug: "northcentralus"},
		{Name: "South Central US", Slug: "southcentralus"},
		{Name: "West Central US", Slug: "westcentralus"},
		{Name: "West US", Slug: "westus"},
		{Name: "West US 2", Slug: "westus2"},
	}
)
