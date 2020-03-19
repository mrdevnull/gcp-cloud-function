package gcpfunction

import (
	"cloud.google.com/go/cloudbuild/apiv1"
	"context"
	"encoding/json"
	"fmt"
	cloudbuildpb "google.golang.org/genproto/googleapis/devtools/cloudbuild/v1"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type ContainerRegistryTriggerMessage struct {
	Action string
	Digest string
	Tag    string
}

func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))

	var crtm ContainerRegistryTriggerMessage

	_ = json.Unmarshal([]byte(m.Data), &crtm)

	fmt.Println("Action: ", crtm.Action)
	fmt.Println("Digest: ", crtm.Digest)
	fmt.Println("Tag: ", crtm.Tag)

	ctx := context.Background()
	c, err := cloudbuild.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	buildSteps := cloudbuildpb.BuildStep{
		Name:                 "",
		Env:                  nil,
		Args:                 nil,
		Dir:                  "",
		Id:                   "",
		WaitFor:              nil,
		Entrypoint:           "",
		SecretEnv:            nil,
		Volumes:              nil,
		Timing:               nil,
		PullTiming:           nil,
		Timeout:              nil,
		Status:               0,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}

	build := cloudbuildpb.Build{
		Id:                   "$BUILD_ID",
		ProjectId:            "$PROJECT_ID",
		Status:               0,
		StatusDetail:         "",
		Source:               nil,
		Steps:                &buildSteps,
		Results:              nil,
		CreateTime:           nil,
		StartTime:            nil,
		FinishTime:           nil,
		Timeout:              nil,
		Images:               nil,
		Artifacts:            nil,
		LogsBucket:           "",
		SourceProvenance:     nil,
		BuildTriggerId:       "",
		Options:              nil,
		LogUrl:               "",
		Substitutions:        nil,
		Tags:                 nil,
		Secrets:              nil,
		Timing:               nil,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}

	req := &cloudbuildpb.CreateBuildRequest{
		ProjectId:            "goodbetterbest",
		Build:                &build,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}

	resp, err := c.CreateBuild(ctx, req)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// TODO: Use resp.
	_ = resp

	cloudbuildClient := http.Client{
		Timeout: time.Second * 3,
	}

	cloudBuildStepsTemplate := fmt.Sprintf(
		`
		{
			"steps": [
				{
					"name":"gcr.io/cloud-builders/gke-deploy",
					"args":[
						"run",
						"--filename=gs://go-hello-world/application.yml",
						"--location=australia-southeast1-a","--cluster=cluster",
						"--image=%s"
					]
				}
			]
		}
		`,
		crtm.Tag,
	)

	log.Println(cloudBuildStepsTemplate)

	cloudBuildSteps := strings.NewReader(cloudBuildStepsTemplate)

	req, err := http.NewRequest(http.MethodPost, "https://cloudbuild.googleapis.com/v1/projects/goodbetterbest/builds", cloudBuildSteps)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	res, err2 := cloudbuildClient.Do(req)
	if err2 != nil {
		log.Fatal(err2)
		return nil
	}

	body, err3 := ioutil.ReadAll(res.Body)
	if err3 != nil {
		log.Fatal(err3)
	}

	log.Println(string(body))

	log.Println("Finished codedeploy initiation.")

	return nil
}
