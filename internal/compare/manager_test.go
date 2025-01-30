// File: internal/compare/manager_test.go

package compare

import (
	"fmt"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
)

// MockResource is a simple implementation of runtime.Object for testing
type MockResource struct {
	Name      string
	Kind      string
	Namespace string
}

func (m *MockResource) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (m *MockResource) DeepCopyObject() runtime.Object {
	return &MockResource{Name: m.Name, Kind: m.Kind, Namespace: m.Namespace}
}

// MockLogger is a simple logger for testing
type MockLogger struct {
	InfoMessages  []string
	ErrorMessages []string
}

func (m *MockLogger) Info(args ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, fmt.Sprint(args...))
}
func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, fmt.Sprintf(format, args...))
}
func (m *MockLogger) Error(args ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, fmt.Sprint(args...))
}
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, fmt.Sprintf(format, args...))
}

func TestCompareResources(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	mockClient := &kubernetes.Client{Clientset: fakeClientset}
	mockLogger := &MockLogger{}
	manager := NewManager(mockClient, mockLogger)

	source := []runtime.Object{
		&MockResource{Name: "resource1", Kind: "Deployment", Namespace: "default"},
		&MockResource{Name: "resource2", Kind: "Service", Namespace: "default"},
	}
	target := []runtime.Object{
		&MockResource{Name: "resource1", Kind: "Deployment", Namespace: "default"},
		&MockResource{Name: "resource3", Kind: "ConfigMap", Namespace: "default"},
	}

	differences := manager.compareResources(source, target)

	if len(differences) != 2 {
		t.Errorf("Expected 2 differences, got %d", len(differences))
	}

	for _, diff := range differences {
		switch diff.Type {
		case "Extra":
			if diff.Object.(*MockResource).Name != "resource2" {
				t.Errorf("Expected extra resource 'resource2', got '%s'", diff.Object.(*MockResource).Name)
			}
		case "Missing":
			if diff.Object.(*MockResource).Name != "resource3" {
				t.Errorf("Expected missing resource 'resource3', got '%s'", diff.Object.(*MockResource).Name)
			}
		default:
			t.Errorf("Unexpected difference type: %s", diff.Type)
		}
	}
}

func TestGetObjectKey(t *testing.T) {
	resource := &MockResource{Name: "testResource", Kind: "TestKind", Namespace: "default"}
	key := getObjectKey(resource)
	expected := "TestKind/default/testResource"
	if key != expected {
		t.Errorf("Expected key '%s', got '%s'", expected, key)
	}
}

func TestPerformCompare(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	mockClient := &kubernetes.Client{Clientset: fakeClientset}
	mockLogger := &MockLogger{}
	manager := NewManager(mockClient, mockLogger)

	err := manager.PerformCompare("source", "target", "all", true)

	if err == nil {
		t.Error("Expected an error due to unimplemented features, but got nil")
	}

	// Check if logger received expected messages
	if len(mockLogger.InfoMessages) < 1 {
		t.Errorf("Expected at least 1 info message, got %d", len(mockLogger.InfoMessages))
	}

	// Check for specific log messages
	expectedMsg := "Comparing all resources from source to target"
	if mockLogger.InfoMessages[0] != expectedMsg {
		t.Errorf("Expected log message '%s', got '%s'", expectedMsg, mockLogger.InfoMessages[0])
	}
}
