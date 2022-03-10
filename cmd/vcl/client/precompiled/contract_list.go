package precompile

import "github.com/Venachain/Venachain/common/syscontracts"

var (
	UserManagementAddress        = syscontracts.UserManagementAddress.String()        // The Venachain Precompiled contract addr for user management
	NodeManagementAddress        = syscontracts.NodeManagementAddress.String()        // The Venachain Precompiled contract addr for node management
	CnsManagementAddress         = syscontracts.CnsManagementAddress.String()         // The Venachain Precompiled contract addr for CNS
	ParameterManagementAddress   = syscontracts.ParameterManagementAddress.String()   // The Venachain Precompiled contract addr for parameter management
	FirewallManagementAddress    = syscontracts.FirewallManagementAddress.String()    // The Venachain Precompiled contract addr for fire wall management
	GroupManagementAddress       = syscontracts.GroupManagementAddress.String()       // The Venachain Precompiled contract addr for group management
	ContractDataProcessorAddress = syscontracts.ContractDataProcessorAddress.String() // The Venachain Precompiled contract addr for group management
	CnsInvokeAddress             = syscontracts.CnsInvokeAddress.String()             // The Venachain Precompiled contract addr for group management
)

const (
	PermDeniedEvent = "the contract deployment is denied"
	CnsInvokeEvent  = "the event generated by cns Invoke"
	CnsInitRegEvent = "register the contract to cns from init()"
)

// link the precompiled contract addresses with abi file bytes
var List = map[string]string{
	UserManagementAddress:        "../../release/linux/conf/contracts/userManager.cpp.abi.json",
	NodeManagementAddress:        "../../release/linux/conf/contracts/nodeManager.cpp.abi.json",
	CnsManagementAddress:         "../../release/linux/conf/contracts/cnsManager.cpp.abi.json",
	ParameterManagementAddress:   "../../release/linux/conf/contracts/paramManager.cpp.abi.json",
	FirewallManagementAddress:    "../../release/linux/conf/contracts/fireWall.abi.json",
	GroupManagementAddress:       "../../release/linux/conf/contracts/groupManager.cpp.abi.json",
	ContractDataProcessorAddress: "../../release/linux/conf/contracts/contractData.cpp.abi.json",

	CnsInitRegEvent: "../../release/linux/conf/contracts/cnsInitRegEvent.json",
	CnsInvokeEvent:  "../../release/linux/conf/contracts/cnsInvokeEvent.json",
	PermDeniedEvent: "../../release/linux/conf/contracts/permissionDeniedEvent.json",
}
