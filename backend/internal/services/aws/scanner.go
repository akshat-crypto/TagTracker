package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/sts"
)

var (
	ec2Instances []EC2Instance
	rdsInstances []RDSInstance
)

type EC2Instance struct {
	InstanceID   string
	InstanceType string
	State        string
	Tags         map[string]string
}

type RDSInstance struct {
	DBInstanceIdentifier string
	DBInstanceClass      string
	Engine               string
	Tags                 map[string]string
}

func GetEC2Instances() []EC2Instance {
	return ec2Instances
}

func GetRDSInstances() []RDSInstance {
	return rdsInstances
}

func StartScan(roleARN string) error {
	//TODO: Change approach: Requires environment to have client creds to run the sts assume policy
	//TODO: There will be two different sessions with this one single code, how can we avoid this situation
	// Create an initial session from the environment (EC2 instance, ECS task, Lambda, etc.)
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Assume the role to get temporary credentials
	svc := sts.New(sess)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("TagScannerSession"),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		return fmt.Errorf("failed to assume role: %w", err)
	}

	// Create a new session using the temporary credentials
	creds := credentials.NewStaticCredentials(
		*result.Credentials.AccessKeyId,
		*result.Credentials.SecretAccessKey,
		*result.Credentials.SessionToken,
	)

	tempSess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})
	if err != nil {
		return fmt.Errorf("failed to create session with assumed role: %w", err)
	}

	fmt.Println("Assumed role successfully:", result)

	// Scan EC2 Instances
	ec2Instances, err = ScanEC2Instances(tempSess)
	if err != nil {
		return fmt.Errorf("failed to scan EC2 instances: %w", err)
	}

	// Scan RDS Instances
	rdsInstances, err = ScanRDSInstances(tempSess)
	if err != nil {
		return fmt.Errorf("failed to scan RDS instances: %w", err)
	}

	return nil

}

func ScanEC2Instances(sess *session.Session) ([]EC2Instance, error) {
	svc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []EC2Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			tags := make(map[string]string)
			for _, tag := range instance.Tags {
				tags[*tag.Key] = *tag.Value
			}
			instances = append(instances, EC2Instance{
				InstanceID:   *instance.InstanceId,
				InstanceType: *instance.InstanceType,
				State:        *instance.State.Name,
				Tags:         tags,
			})
		}
	}

	return instances, nil
}

func ScanRDSInstances(sess *session.Session) ([]RDSInstance, error) {
	svc := rds.New(sess)
	input := &rds.DescribeDBInstancesInput{}

	result, err := svc.DescribeDBInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	var instances []RDSInstance
	for _, dbInstance := range result.DBInstances {
		tags := make(map[string]string)
		instances = append(instances, RDSInstance{
			DBInstanceIdentifier: *dbInstance.DBInstanceIdentifier,
			DBInstanceClass:      *dbInstance.DBInstanceClass,
			Engine:               *dbInstance.Engine,
			Tags:                 tags,
		})
	}

	return instances, nil
}
