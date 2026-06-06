package api

import (
	"slices"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func CopyTo(dst protoreflect.Message, src protoreflect.Message, forceSet bool, ignore ...string) {
	dstFields := dst.Descriptor().Fields()
	srcFields := src.Descriptor().Fields()
	for i := range dstFields.Len() {
		dstF := dstFields.Get(i)
		if slices.Contains(ignore, string(dstF.Name())) {
			continue
		}
		srcF := srcFields.ByName(dstF.Name())
		if dstF == nil || srcF == nil {
			continue
		}
		if forceSet || src.Has(srcF) {
			dst.Set(dstF, src.Get(srcF))
		}
	}
}

type GenericRespList[T any] interface {
	HasPaginationToken() bool
	GetPaginationToken() string

	GetResult() []T
	GetTotalCount() uint32
}

func ListAll[T any](fetch func(tok *string) (GenericRespList[T], error)) ([]T, error) {
	resp, err := fetch(nil)
	if err != nil {
		return nil, err
	}

	res := make([]T, resp.GetTotalCount())
	copied := copy(res, resp.GetResult())
	if !resp.HasPaginationToken() {
		return res, nil
	}

	var tok *string = new(resp.GetPaginationToken())

	for {
		resp, err := fetch(tok)
		if err != nil {
			return res, err
		}
		copied += copy(res[copied:], resp.GetResult())
		if !resp.HasPaginationToken() {
			return res, nil
		}

		tok = new(resp.GetPaginationToken())
	}
}
