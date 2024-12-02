package controller

import (
	"context"
	"fmt"
	demov1 "spotpool/api/v1"
	"time"

	logger "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var (
	errInstanceRetry = fmt.Errorf("retry after 40")
)

func (r *SpotPoolReconciler) checkRunningInstances(ctx context.Context, spotpool *demov1.SpotPool) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// 获取 Running 实例数量，这里需要一直检查，因为 IP 地址是异步分配的
	_, err := r.getRunningInstanceIDs(spotpool)
	if err != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, errInstanceRetry
	}

	// Get the current number of running CVM instances.
	runningCount, err := r.countRunningInstances(spotpool)
	if err != nil {
		log.Error(err, "failed to get running CVM instance count")
		return ctrl.Result{}, err
	}

	// Take actions based on the desired state (Minimum and Maximum) and current state (runningCount).
	switch {
	case runningCount < spotpool.Spec.Minimum:
		// Need to start more CVMs.
		delta := spotpool.Spec.Minimum - runningCount
		logger.Info("Starting ", delta, " CVM instances")
		err = r.runInstances(spotpool, delta)
		if err != nil {
			log.Info(err.Error(), "failed to start CVM instances")
			// 40s 后重试
			return ctrl.Result{RequeueAfter: 40 * time.Second}, errInstanceRetry
		}
	case runningCount > spotpool.Spec.Maximum:
		// Need to terminate CVMs.
		delta := runningCount - spotpool.Spec.Maximum
		logger.Info("Terminating ", delta, " CVM instances")
		err = r.terminateInstances(spotpool, delta)
		if err != nil {
			log.Info(err.Error(), "failed to terminate CVM instances")
			return ctrl.Result{RequeueAfter: 40 * time.Second}, errInstanceRetry
		}
	}

	// Update the Spotpool status with the current number of running instances.
	spotpool.Status.Size = runningCount
	err = r.Status().Update(ctx, spotpool)
	if err != nil {
		log.Error(err, "failed to update Spotpool status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// getRunningInstanceIDs gets the IDs of running CVM instances that match the Spotpool configuration.
func (r *SpotPoolReconciler) getRunningInstanceIDs(spotpool *demov1.SpotPool) ([]string, error) {
	// Implement your logic here to query the Tencent Cloud CVM API to get the IDs of running CVM instances that match the Spotpool configuration.
	// This example code assumes you have a functional CVM client and implement a basic search based on InstanceIds from the Spotpool spec. You might need to adjust the logic based on your actual implementation.
	client, err := r.createCVMClient(spotpool.Spec)
	if err != nil {
		return nil, err
	}
	request := cvm.NewDescribeInstancesRequest()
	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	var instances []demov1.Instances
	var runningInstanceIDs []string
	for _, instance := range response.Response.InstanceSet {
		if *instance.InstanceState == "RUNNING" || *instance.InstanceState == "PENDING" || *instance.InstanceState == "STARTING" {
			runningInstanceIDs = append(runningInstanceIDs, *instance.InstanceId)
		}
		// 检查实例的公网 IP，如果不存在公网 IP 则返回错误，下次重试
		if len(instance.PublicIpAddresses) == 0 {
			return nil, fmt.Errorf("instance %s does not have a public IP", *instance.InstanceId)
		}
		instances = append(instances, demov1.Instances{
			InstanceId: *instance.InstanceId,
			PublicIp:   *instance.PublicIpAddresses[0],
		})
	}
	// update status with runningInstanceIDs
	spotpool.Status.Instances = instances
	err = r.Status().Update(context.Background(), spotpool)
	if err != nil {
		return nil, err
	}
	return runningInstanceIDs, nil
}

// countRunningInstances gets the number of CVMs running in the specified Spotpool's configuration.
func (r *SpotPoolReconciler) countRunningInstances(spotpool *demov1.SpotPool) (int32, error) {
	// Implement your logic here to query the Tencent Cloud CVM API and return the number of running instances that match the Spotpool configuration (e.g., Region, InstanceType, etc.)
	// This example code assumes you have a functional CVM client and implement a basic search based on InstanceIds from the Spotpool spec. You might need to adjust the logic based on your actual implementation.
	client, err := r.createCVMClient(spotpool.Spec)
	if err != nil {
		return 0, err
	}
	request := cvm.NewDescribeInstancesRequest()
	response, err := client.DescribeInstances(request)
	if err != nil {
		return 0, err
	}
	var runningCount int32
	for _, instance := range response.Response.InstanceSet {
		logger.Info("Instance ", *instance.InstanceId, " State ", *instance.InstanceState)
		if *instance.InstanceState == "RUNNING" || *instance.InstanceState == "PENDING" || *instance.InstanceState == "STARTING" {
			runningCount++
		}
	}
	logger.Info("Found ", runningCount, " running instances")
	return runningCount, nil
}

// runInstances starts the specified number of CVMs based on the Spotpool configuration.
func (r *SpotPoolReconciler) runInstances(spotpool *demov1.SpotPool, count int32) error {
	// Implement your logic here to call the Tencent Cloud CVM API to start new CVM instances based on the Spotpool spec.
	client, err := r.createCVMClient(spotpool.Spec)
	if err != nil {
		return err
	}

	request := cvm.NewRunInstancesRequest()
	request.ImageId = common.StringPtr(spotpool.Spec.ImageId)
	request.Placement = &cvm.Placement{
		Zone: common.StringPtr(spotpool.Spec.AvailabilityZone),
	}

	request.InstanceChargeType = common.StringPtr(spotpool.Spec.InstanceChargeType)
	request.InstanceCount = common.Int64Ptr(int64(count))
	request.InstanceName = common.StringPtr("spotpool-instance-" + time.Now().Format("20060102150405"))
	request.InstanceType = common.StringPtr(spotpool.Spec.InstanceType)
	request.InternetAccessible = &cvm.InternetAccessible{
		InternetChargeType:      common.StringPtr("BANDWIDTH_POSTPAID_BY_HOUR"),
		InternetMaxBandwidthOut: common.Int64Ptr(100),
		PublicIpAssigned:        common.BoolPtr(true),
	}
	request.LoginSettings = &cvm.LoginSettings{
		Password: common.StringPtr("Password123"),
	}
	request.SecurityGroupIds = common.StringPtrs(spotpool.Spec.SecurityGroupIds)
	request.SystemDisk = &cvm.SystemDisk{
		DiskSize: common.Int64Ptr(100),
		DiskType: common.StringPtr("CLOUD_BSSD"),
	}
	request.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		SubnetId: common.StringPtr(spotpool.Spec.SubnetId),
		VpcId:    common.StringPtr(spotpool.Spec.VpcId),
	}
	// cvm instance number
	request.InstanceCount = common.Int64Ptr(int64(count))

	// get response structure
	response, err := client.RunInstances(request)
	// API errors
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}
	// unexpected errors
	if err != nil {
		panic(err)
	}
	// Extract instance IDs from the response
	instanceIds := make([]string, 0, len(response.Response.InstanceIdSet))
	for _, instanceId := range response.Response.InstanceIdSet {
		instanceIds = append(instanceIds, *instanceId)
	}
	logger.Info("Started ", len(instanceIds), " CVM instances: ", instanceIds)
	// Get the instance IDs to terminate, for update status
	_, err = r.getRunningInstanceIDs(spotpool)
	if err != nil {
		return err
	}
	return nil
}

// terminateInstances terminates the specified number of CVMs based on the Spotpool configuration.
func (r *SpotPoolReconciler) terminateInstances(spotpool *demov1.SpotPool, count int32) error {
	// Implement your logic here to call the Tencent Cloud CVM API to terminate existing CVM instances based on the Spotpool spec.
	// This example code shows a basic invocation but you would need to fill in the details based on the actual API call.
	client, err := r.createCVMClient(spotpool.Spec)
	if err != nil {
		return err
	}

	// Get the instance IDs to terminate
	runningInstances, err := r.getRunningInstanceIDs(spotpool)
	if err != nil {
		return err
	}
	// Select the instance IDs to terminate based on count
	instanceIds := runningInstances[:count]
	logger.Info("Terminating", len(instanceIds), "CVM instances:", instanceIds)

	request := cvm.NewTerminateInstancesRequest()
	request.InstanceIds = common.StringPtrs(instanceIds)

	// get response structure
	response, err := client.TerminateInstances(request)
	// API errors
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Log.Error(err, "failed to terminate CVM instances")
		return err
	}
	// unexpected errors
	if err != nil {
		log.Log.Error(err, "failed to terminate CVM instances")
	}
	logger.Info("Response", response)
	logger.Info("Terminated", len(instanceIds), "CVM instances")

	// Get the instance IDs to terminate, for update status
	_, err = r.getRunningInstanceIDs(spotpool)
	if err != nil {
		return err
	}

	return nil
}

// createCVMClient creates a Tencent Cloud CVM client based on the Spotpool spec.
func (r *SpotPoolReconciler) createCVMClient(spec demov1.SpotPoolSpec) (*cvm.Client, error) {
	credential := common.NewCredential(spec.SecretId, spec.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 30
	//cpf.HttpProfile.Endpoint = "cvm.ap-guangzhou.tencentcloudapi.com"
	cpf.SignMethod = "HmacSHA1"

	client, err := cvm.NewClient(credential, spec.Region, cpf)
	if err != nil {
		return nil, err
	}
	return client, nil
}
