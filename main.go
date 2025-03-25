package main

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	clientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	informers "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
	"k8s.io/client-go/tools/cache"
)

func main() {
	kubeconfig := "/Users/rokumar/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}
	// Use in-cluster configuration for connecting to the Kubernetes API
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	log.Fatalf("Error creating in-cluster config: %v", err)
	//}

	// Create Tekton client
	tektonClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Tekton client: %v", err)
	}

	// Create dynamic client for TestSuiteRun
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	// Define the GVR (Group-Version-Resource) for TestSuiteRun
	testSuiteRunGVR := schema.GroupVersionResource{
		Group:    "example.com",
		Version:  "v1",
		Resource: "testsuiteruns",
	}

	// Create Informer using a List-Watch function
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				// Use dynamic client to list TestSuiteRuns
				unstructuredList, err := dynamicClient.Resource(testSuiteRunGVR).Namespace("default").List(context.TODO(), options)
				return unstructuredList, err
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				// Use dynamic client to watch TestSuiteRuns
				return dynamicClient.Resource(testSuiteRunGVR).Namespace("default").Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{}, // Unstructured type for dynamic client
		0,                            // No resync period
		cache.Indexers{},
	)

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Handle the creation of a new TestSuiteRun
			testSuiteRun := obj.(*unstructured.Unstructured)
			name := testSuiteRun.GetName()
			log.Printf("TestSuiteRun created: %s", name)

			// Extract fields from TestSuiteRun (machineType and testName)
			spec, found, err := unstructured.NestedMap(testSuiteRun.Object, "spec")
			if !found || err != nil {
				log.Printf("Error extracting spec from TestSuiteRun: %v", err)
				return
			}
			machineType := spec["machineType"].(string)
			testName := spec["testName"].(string)

			// Create a PipelineRun for the TestSuiteRun
			pipelineRun := &pipelinev1.PipelineRun{
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("%s-pipelinerun", name),
					Namespace: "default",
					Labels: map[string]string{
						"machine-type": machineType,
						"testname":     testName,
						"owner":        name,
					},
				},
				Spec: pipelinev1.PipelineRunSpec{
					PipelineSpec: &pipelinev1.PipelineSpec{
						Tasks: []pipelinev1.PipelineTask{
							{
								Name: "example-task",
								TaskSpec: &pipelinev1.EmbeddedTask{
									TaskSpec: pipelinev1.TaskSpec{
										Steps: []pipelinev1.Step{},
									},
								},
							},
						},
					},
				},
			}

			// Create PipelineRun in the cluster
			_, err = tektonClient.TektonV1().PipelineRuns("default").Create(context.TODO(), pipelineRun, v1.CreateOptions{})
			if err != nil {
				log.Printf("Error creating PipelineRun: %v", err)
				return
			}
			log.Printf("PipelineRun %s created for TestSuiteRun %s", pipelineRun.Name, name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Handle updates if necessary
			log.Println("TestSuiteRun updated")
		},
		DeleteFunc: func(obj interface{}) {
			// Handle deletions if necessary
			log.Println("TestSuiteRun deleted")
		},
	})

	// Create shared informer factory
	factory := informers.NewSharedInformerFactory(tektonClient, time.Minute*10)

	// Get PipelineRun informer
	pipelineRunInformer := factory.Tekton().V1().PipelineRuns().Informer()

	_, err = pipelineRunInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pr := obj.(*pipelinev1.PipelineRun)
			fmt.Printf("PipelineRun added: %s\n", pr.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPR := newObj.(*pipelinev1.PipelineRun)
			fmt.Printf("PipelineRun updated: %s\n", newPR.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pr := obj.(*pipelinev1.PipelineRun)
			fmt.Printf("PipelineRun deleted: %s\n", pr.Name)
		},
	})
	if err != nil {
		log.Fatalf("Error adding pipeline run event handler: %v", err)
	}

	// Start informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	// Wait for the caches to sync
	if !cache.WaitForCacheSync(stopCh, pipelineRunInformer.HasSynced) {
		log.Fatalf("Failed to sync PipelineRun informer cache")
	}

	log.Println("Tekton operator is running...")
	<-stopCh
}
