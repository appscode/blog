---
title: Building Testable Custom Kubernetes Controllers with Golang
date: "2023-06-05"
weight: 26
authors:
- Hossain Mahmud
tags:
- golang
- kubebuilder
- kubernetes
- unittest
---

## Building Testable Custom Kubernetes Controllers with Golang

### Introduction
In the world of Kubernetes, custom controllers play a crucial role in extending the functionality of the platform to meet specific application requirements. However, developing reliable and testable controllers can be challenging. In this blog post, we will explore how to build testable custom Kubernetes controllers using Golang and KubeBuilder. We'll walk through the process step by step, providing code examples and explanations along the way.

### Prerequisites
Before we begin, make sure you have a basic understanding of Golang, Kubernetes, and KubeBuilder. Familiarize yourself with the concept of unit testing and how it applies to Golang applications.

#### Setting Up the Controller:
To begin, let's set up a basic custom controller KubeBuilder. Here's a simplified version of the code:

```go
type DependencyInterface interface {
    DoSomething(client client.Client) string
}

type Reconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

type reconciler struct {
    client     client.Client
    dependency DependencyInterface
}

type RealImplementation struct{}

func (m *RealImplementation) DoSomething(client client.Client) error {
    // Implementation logic goes here
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    r := reconciler{
        client:     r.Client,
        dependency: RealImplementation{},
    }
    if err := r.reconcile(); err != nil {
        return ctrl.Result{}, err
    }
    return ctrl.Result{}, nil
}

func (r *reconciler) reconcile() error {
    // Call the dependent method using the interface
    result := r.dependency.DoSomething(r.client)
}

```
In this code snippet, we have a custom controller, `Reconciler`, that performs reconciliation logic in the `Reconcile` method. The Reconcile method creates an instance of the `reconciler` struct and calls its `reconcile` method.  The Reconcile method creates an instance of the `reconciler` struct and calls its `reconcile` method.

Writing Unit Tests:
o ensure the reliability and correctness of our custom controller, we need to write unit tests. Let's demonstrate how to write a test for the reconcile method.

```go
type MockImplementation struct {
    doSomethingResponse error
}

func (m *MockImplementation) DoSomething(client client.Client) error {
    // Implementation logic goes here
}

func TestReconcile(t *testing.T) {
    fakeClient, err :=  getFakeClient()
    assert.Nil(t, err)
    
    mi := &MockImplementation{
        doSomethingResponse: fmt.Errorf("mock error"),
    }

    reconciler := &reconciler{
        client:     fakeClient,
        dependency: mi,
    }

    err := reconciler.reconcile()
    assert.NotNil(t, err)
}

func getFakeClient(initObjs ...client.Object) (client.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
    // ...
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build(), nil
}

```

In this unit test, we create a mock implementation of the `DependencyInterface` named `MockImplementation`. By implementing the `DoSomething` method, we can control the behavior and responses during the test. In our example, we intentionally return an error to simulate an error scenario.

The test creates a reconciler instance with a fake client and a mock dependency. We then invoke the reconcile method and assert that the returned error is not nil.

By using a mock implementation, we isolate the unit under test and ensure that it behaves as expected, regardless of the external dependencies. This approach allows us to focus solely on testing the logic within the controller.

### Conclusion:
Writing unit tests for custom Kubernetes controllers is essential to ensure their reliability and correctness. By employing Golang and KubeBuilder, we can easily create testable controllers and use mock implementations to isolate dependencies during testing. In this blog post, we demonstrated how to build a testable custom controller and provided a sample unit
## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).



