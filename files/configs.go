package files

import (
	"fmt"
	pb "github.com/omecodes/store/gen/go/proto"
)

var accessNamespaceConfig = pb.NamespaceConfig{
	Sid:       0,
	Namespace: accessNamespace,
	Relations: map[string]*pb.RelationDefinition{
		relationOwner: {
			Name: relationOwner,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
			},
		},
		relationViewer: {
			Name: relationViewer,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationOwner,
				},
			},
		},
	},
}

var groupNamespaceConfig = pb.NamespaceConfig{
	Sid:       0,
	Namespace: accessNamespace,
	Relations: map[string]*pb.RelationDefinition{
		relationOwner: {
			Name: relationOwner,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
			},
		},
		relationMember: {
			Name: relationMember,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationOwner,
				},
			},
		},
	},
}

var fileNamespaceConfig = pb.NamespaceConfig{
	Sid:       0,
	Namespace: fileNamespace,
	Relations: map[string]*pb.RelationDefinition{
		relationOwner: {
			Name: relationOwner,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_FromTuple,
					Value: fmt.Sprintf("{\"object_relation\":  \"%s\", \"subject_relation\":  \"%s\"}", relationParent, relationOwner),
				},
			},
		},
		relationEditor: {
			Name: relationEditor,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationOwner,
				},
				{
					Type:  pb.SubjectSetType_FromTuple,
					Value: fmt.Sprintf("{\"object_relation\":  \"%s\", \"subject_relation\":  \"%s\"}", relationParent, relationEditor),
				},
			},
		},
		relationSharer: {
			Name: relationSharer,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationOwner,
				},
				{
					Type:  pb.SubjectSetType_FromTuple,
					Value: fmt.Sprintf("{\"object_relation\":  \"%s\", \"subject_relation\":  \"%s\"}", relationParent, relationSharer),
				},
			},
		},
		relationViewer: {
			Name: relationViewer,
			SubjectSetRewrite: []*pb.SubjectSetDefinition{
				{
					Type: pb.SubjectSetType_This,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationOwner,
				},
				{
					Type:  pb.SubjectSetType_Computed,
					Value: relationEditor,
				},
				{
					Type:  pb.SubjectSetType_FromTuple,
					Value: fmt.Sprintf("{\"object_relation\":  \"%s\", \"subject_relation\":  \"%s\"}", relationParent, relationViewer),
				},
			},
		},
	},
}
