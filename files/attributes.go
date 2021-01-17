package files

import "github.com/omecodes/store/pb"

type Attributes map[AttrName]string

type AttrName string

type Attribute struct {
	Name  AttrName
	Value []byte
}

const (
	AttrPathHistory      = AttrName("store-path-history")
	AttrSize             = AttrName("store-size")
	AttrCreatedBy        = AttrName("store-created-by")
	AttrReadPermissions  = AttrName("store-read-permissions")
	AttrWritePermissions = AttrName("store-write-permissions")
	AttrChmodPermissions = AttrName("store-chmod-permissions")
)

func DecodePermissions(attrValue string) ([]string, error) {
	return nil, nil
}

func DecodePathHistory(attrValue string) ([]string, error) {
	return nil, nil
}

func SizeFromMeta(attrValue string) (int64, error) {
	return 0, nil
}

func CreatorFromMeta(attrValue string) (string, error) {
	return "", nil
}

func SetAttributes(filename string, attrs ...Attribute) error {
	return nil
}

func SetPermissions(filename string, rules *pb.FileAccessRules) error {
	return nil
}

func AppendToPathHistoryAttribute(filename string, paths ...string) error {
	return nil
}
