package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/akshat-crypto/TagTracker/backend/internal/db"
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
	InstanceID   string            `json:"instance_id"`
	InstanceType string            `json:"instance_type"`
	State        string            `json:"state"`
	Tags         map[string]string `json:"tags"`
	TimeStamp    time.Time         `json:"created_at"`
}

type RDSInstance struct {
	DBInstanceIdentifier string            `json:"db_instance_identifier"`
	DBInstanceClass      string            `json:"db_instance_class"`
	Engine               string            `json:"engine"`
	Tags                 map[string]string `json:"tags"`
	TimeStamp            time.Time         `json:"created_at"`
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

	var wg sync.WaitGroup

	wg.Add(2) //Running two concurrent goroutines

	go func() {
		defer wg.Done()
		ec2Instances, err = ScanEC2Instances(tempSess)
		if err != nil {
			log.Printf("failed to scan EC2 instances: %v", err)
			return
		}
		storeEC2Instance(ec2Instances)
	}()

	go func() {
		defer wg.Done()
		rdsInstances, err = ScanRDSInstances(tempSess)
		if err != nil {
			log.Printf("failed to scan RDS instances: %v", err)
			return
		}
		storeRDSInstance(rdsInstances)
	}()

	wg.Wait()

	return nil

}

// TODO: Separate files for scanning or storing services
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
				TimeStamp:    *instance.LaunchTime,
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
			TimeStamp:            *dbInstance.InstanceCreateTime,
			Tags:                 tags,
		})
	}

	return instances, nil
}

// TODO: Change this function to separate file
func storeEC2Instance(instances []EC2Instance) error {

	db := db.GetDB()

	for _, instance := range instances {
		tags, _ := json.Marshal(instance.Tags)
		_, err := db.Exec(
			`INSERT INTO ec2_instances (instance_id, instance_type, state, tags, created_at) 
             VALUES ($1, $2, $3, $4, $5)
             ON CONFLICT (instance_id) DO UPDATE 
             SET instance_type = EXCLUDED.instance_type, state = EXCLUDED.state, tags = EXCLUDED.tags, created_at = EXCLUDED.created_at`,
			instance.InstanceID, instance.InstanceType, instance.State, tags, instance.TimeStamp,
		)
		if err != nil {
			log.Printf("Failed to insert EC2 instance: %v", err)
			return err
		}
	}

	return nil
}

func storeRDSInstance(instances []RDSInstance) error {

	db := db.GetDB()

	for _, instance := range instances {
		tags, _ := json.Marshal(instance.Tags)
		_, err := db.Exec(
			`INSERT INTO rds_instances (db_instance_identifier, db_instance_class, engine, tags, created_at) 
             VALUES ($1, $2, $3, $4, $5)
             ON CONFLICT (db_instance_identifier) DO UPDATE 
             SET db_instance_class = EXCLUDED.db_instance_class, engine = EXCLUDED.engine, tags = EXCLUDED.tags, created_at = EXCLUDED.created_at`,
			instance.DBInstanceIdentifier, instance.DBInstanceClass, instance.Engine, tags, instance.TimeStamp,
		)
		if err != nil {
			log.Printf("Failed to insert RDS instance: %v", err)
			return err
		}
	}

	return nil
}

//TODO: Fetch EC2 instance resources

func FetchEC2Instances() ([]EC2Instance, error) {
	db := db.GetDB()
	rows, err := db.Query("SELECT instance_id, instance_type, state, tags, created_at FROM ec2_instances")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []EC2Instance
	for rows.Next() {
		var instance EC2Instance
		var tags []byte
		if err := rows.Scan(&instance.InstanceID, &instance.InstanceType, &instance.State, &tags, &instance.TimeStamp); err != nil {
			return nil, err
		}
		json.Unmarshal(tags, &instance.Tags)
		instances = append(instances, instance)
	}

	return instances, nil
}

func FetchRDSInstances() ([]RDSInstance, error) {
	db := db.GetDB()

	rows, err := db.Query("SELECT db_instance_identifier, db_instance_class, engine, tags, created_at FROM rds_instances")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var instances []RDSInstance

	for rows.Next() {
		var instance RDSInstance
		var tags []byte
		if err := rows.Scan(&instance.DBInstanceIdentifier, &instance.DBInstanceClass, &instance.Engine, &tags, &instance.TimeStamp); err != nil {
			return nil, err
		}
		json.Unmarshal(tags, &instance.Tags)
		instances = append(instances, instance)
	}

	return instances, nil
}
