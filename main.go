package main

import (
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 读取 kubeconfig
	kubeconfig := flag.String("kubeconfig", clientcmd.RecommendedHomeFile, "Path to kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// 使用 dynamic client 操作 CR
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 目标 GVR
	gvr := schema.GroupVersionResource{
		Group:    "example.com",
		Version:  "v1",
		Resource: "helloworlds",
	}

	// 创建 informer
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynClient,
		time.Minute,
		metav1.NamespaceAll,
		nil,
	)

	informer := factory.ForResource(gvr).Informer()

	// 设置事件处理器
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			spec := u.Object["spec"].(map[string]interface{})
			msg := spec["message"].(string)
			fmt.Printf("🎉 New HelloWorld created: %s\n", msg)
		},
	})

	// 启动 informer
	stop := make(chan struct{})
	defer close(stop)
	factory.Start(stop)

	// 等待缓存同步
	if ok := cache.WaitForCacheSync(stop, informer.HasSynced); !ok {
		panic("cache sync failed")
	}

	fmt.Println("🔧 Operator is watching HelloWorld resources...")
	<-stop
}
