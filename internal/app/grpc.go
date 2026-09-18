package app

import idpv1 "github.com/al-revenko/idp/proto/gen/idp"

var PublicRPCs = map[string]struct{}{
	idpv1.ClientService_ClientCreate_FullMethodName: {},
}
