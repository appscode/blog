---
title: Writing Unit Tests for Custom Kubernetes Controllers in Go
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

### Introduction

In the world of Kubernetes, custom controllers play a crucial role in extending the functionality of the platform to meet specific application requirements. However, developing reliable and testable controllers can be challenging. In this blog post, we will explore how to write effective unit tests for custom Kubernetes controllers using Golang. We'll walk through the process step by step, providing code examples and explanations along the way.

### Prerequisites

Before we begin, make sure you have a basic understanding of Golang, Kubernetes, and KubeBuilder. Familiarize yourself with the concept of unit testing and how it applies to Golang applications.

### Setting Up the Controller

To begin, let's set up a basic controller using KubeBuilder. Below is a simplified version of the code:

```go
type Reconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

type Dependency interface {
    DoSomething(client client.Client) error
}

type reconciler struct {
    client     client.Client
    dependency Dependency
}

type RealImplementation struct{}

func (m RealImplementation) DoSomething(client client.Client) error {
    // Implementation logic goes here
    return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    rec := reconciler{
        client:     r.Client,
        dependency: RealImplementation{},
    }
    if err := rec.reconcile(); err != nil {
        return ctrl.Result{}, err
    }
    return ctrl.Result{}, nil
}

func (r *reconciler) reconcile() error {
    // Reconcilliation logic goes here

    // Call the dependent method using the interface
    return r.dependency.DoSomething(r.client)
}

```

The `Dependency` is an interface with a single method `DoSomething` that takes a `client.Client` and returns an error. This interface is implemented by the `RealImplementation` struct.

The `Reconciler` struct has a `Reconcile` method that takes a context and a request. Inside the method, a `reconciler` instance is created and initialized with the `r.Client` value for the `client` field and a new instance of `RealImplementation` for the `dependency` field. The `Reconcile` method calls the `reconcile` method of the `reconciler` struct, which performs reconciliation logic and then calls the `DoSomething` method on the dependency field, passing in the `client.Client`.


The `reconcile` method acts as a helper method that contains the actual reconciliation logic. By separating it from the `Reconcile` method of the Reconciler struct, it becomes easier to write unit tests specifically for the reconciliation logic without needing to invoke the original `Reconcile` method. During unit testing, you can directly instantiate the `reconciler` struct and call the `reconcile` method, passing any necessary dependencies or mocks. This isolation facilitates focused testing of the reconciliation logic itself, independent of the outer `Reconcile` method and its dependencies.

### Writing Unit Tests

To ensure the reliability and correctness of our controller, we need to write unit tests. Let's demonstrate how to write a test for the reconcile method.

```go
type MockImplementation struct {
    doSomethingResponse error
}

func (m MockImplementation) DoSomething(client client.Client) error {
    // Implementation logic goes here
    return m.doSomethingResponse
}

func TestReconcile(t *testing.T) {
    fakeClient, err :=  getFakeClient()
    assert.Nil(t, err)
    
    mi := MockImplementation{
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

In the `TestReconcile` function, we create a fake client with the necessary scheme and any initial objects required for the test. We then instantiate a `MockImplementation` struct, providing a predefined error response to simulate a failure scenario.

Next, we create an instance of the `reconciler` struct, passing the fake client and the mock implementation as dependencies. Finally, we call the `reconcile` method and assert that it returns a non-nil error.

### Conclusion

In this blog post, we explored how to write unit tests for a custom Kubernetes controller using the Go programming language. We focused on testing the reconciler logic and showcased how to use mock implementations to isolate and control dependencies.

Unit testing is crucial in ensuring the correctness and reliability of our code. By following the example provided in this post, you'll be able to write effective unit tests for your own controller, helping you catch bugs early and build more robust and stable applications.

Remember, the code snippets provided here are simplified for illustration purposes, and in a real-world scenario, you may need to adapt them to your specific use case. Happy testing!
 
## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).



