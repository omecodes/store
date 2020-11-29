package router

/*cel.Functions(
	&functions.Overload {
		Operator: "acl",
		Binary: func(l ref.Val, r ref.Val) ref.Val {
			if types.StringType != l.Type() {
				return types.ValOrErr(l, "expect first argument to be string")
			}
			if types.StringType != r.Type() {
				return types.ValOrErr(r, "expect second argument to be string")
			}

			uid := l.Value().(string)
			uri := r.Value().(string)

			accessStore := accessStore(ctx)
			auth := authInfo(*ctx)

			perm, err := accessStore.G(uid, uri, auth.Uid)
			if err != nil {
				log.Error("could not load user permission", log.Err(err))
				return types.DefaultTypeAdapter.NativeToValue(&oms.Perm{})
			}

			if perm == nil {
				perm = &oms.Permission{}
			}

			return types.DefaultTypeAdapter.NativeToValue(&oms.Perm{
				Read:  perm.Actions&oms.AllowedTo_read == oms.AllowedTo_read,
				Write: perm.Actions&oms.AllowedTo_write == oms.AllowedTo_write,
			})
		},
	},
), */
