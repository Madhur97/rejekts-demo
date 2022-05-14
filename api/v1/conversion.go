package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conv "k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	v2 "rejekts-demo/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts v1 objects to v2 objects
func (src *User) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v2.User)
	if err := Convert_v1_User_To_v2_User(src, dst, nil); err != nil {
		return err
	}
	restored := &v2.User{}

	// fetching v2 object from annotation of v1 object
	ok, err := UnmarshalData(src, restored)
	if err != nil {
		return err
	}
	if ok {
		// If v2 is present in annotation
		dst.Spec.PassportDetail = restored.Spec.PassportDetail
	} else {
		// if v2 is not present in annotation
		dst.Spec.PassportDetail = v2.PassportDetail{
			PassportNumber: src.Spec.PassportNumber,
		}
	}

	return nil
}

//ConvertFrom converts v2 object to v1
func (dst *User) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v2.User)
	if err := Convert_v2_User_To_v1_User(src, dst, nil); err != nil {
		return err
	}

	// store v2 object in annotation of v1 object
	if err := MarshalData(src, dst); err != nil {
		return err
	}

	dst.Spec.PassportNumber = src.Spec.PassportDetail.PassportNumber

	return nil
}

func Convert_v2_UserSpec_To_v1_UserSpec(in *v2.UserSpec, out *UserSpec, s conv.Scope) error {
	return autoConvert_v2_UserSpec_To_v1_UserSpec(in, out, s)
}

func Convert_v1_UserSpec_To_v2_UserSpec(in *UserSpec, out *v2.UserSpec, s conv.Scope) error {
	return autoConvert_v1_UserSpec_To_v2_UserSpec(in, out, s)
}

const DataAnnotation = "info.gov.in/conversion-data"

// MarshalData stores the source object as json data in the destination object annotations map.
// It ignores the metadata of the source object.
func MarshalData(src metav1.Object, dst metav1.Object) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(src)
	if err != nil {
		return err
	}
	delete(u, "metadata")

	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	annotations := dst.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[DataAnnotation] = string(data)
	dst.SetAnnotations(annotations)
	return nil
}

// UnmarshalData tries to retrieve the data from the annotation and unmarshals it into the object passed as input.
func UnmarshalData(from metav1.Object, to interface{}) (bool, error) {
	annotations := from.GetAnnotations()
	data, ok := annotations[DataAnnotation]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal([]byte(data), to); err != nil {
		return false, err
	}
	delete(annotations, DataAnnotation)
	from.SetAnnotations(annotations)
	return true, nil
}
