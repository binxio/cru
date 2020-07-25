package ref

import (
	"testing"
)

func TestListReferenceSimple(t *testing.T) {
	simple := []byte(`eu.gcr.io/binxio/paas-monitor:v0.3.1`)
	expect := ContainerImageReferences{*MustNewContainerImageReference(`eu.gcr.io/binxio/paas-monitor:v0.3.1`)}
	result := FindAllContainerImageReference(simple)
	for i, ref := range result {
		if i >= len(expect) || ref.Compare(expect[i]) != 0 {
			t.Errorf("expected %s, got %s\n", expect, result)
		}
	}
}

func TestListReferenceMultiple(t *testing.T) {
	input := []byte(`this is one eu.gcr.io/binxio/paas-monitor:v0.3.1 reference
and this is another mvanholsteijn/paas-monitor:3.1.0
and this is just a directory Name mvanholsteijn/paas-monitor, which should
not be changed`)
	expect := []ContainerImageReference{*MustNewContainerImageReference(`eu.gcr.io/binxio/paas-monitor:v0.3.1`),
		*MustNewContainerImageReference(`mvanholsteijn/paas-monitor:3.1.0`)}

	result := FindAllContainerImageReference(input)
	for i, ref := range result {
		if i >= len(expect) || ref.Compare(expect[i]) != 0 {
			t.Errorf("expected %s, got %s\n", expect, result)
		}
	}
}
