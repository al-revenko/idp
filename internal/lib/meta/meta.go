package meta

import (
	"context"

	"github.com/al-revenko/idp/internal/lib/sign"
)

var pkg = sign.Pkg("meta")

type CtxReqIDKey struct{}

func ExtractReqID(ctx context.Context) string {
	if reqId, ok := ctx.Value(CtxReqIDKey{}).(string); ok {
		return reqId
	}
	return ""
}

type CtxClientIDKey struct{}

func ExtractClientID(ctx context.Context) string {
	if clientId, ok := ctx.Value(CtxClientIDKey{}).(string); ok {
		return clientId
	}
	return ""
}

type CtxRedisTxKey struct{}
