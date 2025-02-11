package sagemaker

import (
	"context"
	"fmt"

	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func createOutputLiteralMap(tk *core.TaskTemplate, outputPath string) *core.LiteralMap {
	op := &core.LiteralMap{}
	for k := range tk.Interface.Outputs.Variables {
		// if v != core.LiteralType_Blob{}
		op.Literals = make(map[string]*core.Literal)
		op.Literals[k] = &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Blob{
						Blob: &core.Blob{
							Metadata: &core.BlobMetadata{
								Type: &core.BlobType{Dimensionality: core.BlobType_SINGLE},
							},
							Uri: outputPath,
						},
					},
				},
			},
		}
	}
	return op
}

func getOutputLiteralMapFromTaskInterface(ctx context.Context, tr pluginsCore.TaskReader, outputPath string) (*flyteIdlCore.LiteralMap, error) {
	tk, err := tr.Read(ctx)
	if err != nil {
		return nil, err
	}
	if tk.Interface.Outputs != nil && tk.Interface.Outputs.Variables == nil {
		logger.Warnf(ctx, "No outputs declared in the output interface. Ignoring the generated outputs.")
		return nil, nil
	}

	// We know that for XGBoost task there is only one output to be generated
	if len(tk.Interface.Outputs.Variables) > 1 {
		return nil, fmt.Errorf("expected to generate more than one outputs of type [%v]", tk.Interface.Outputs.Variables)
	}
	op := createOutputLiteralMap(tk, outputPath)
	return op, nil
}

func createOutputPath(prefix string, subdir string) string {
	return fmt.Sprintf("%s/%s", prefix, subdir)
}

func createModelOutputPath(job client.Object, prefix, jobName string) string {
	switch job.(type) {
	case *trainingjobv1.TrainingJob:
		return fmt.Sprintf("%s/%s/output/model.tar.gz", createOutputPath(prefix, TrainingJobOutputPathSubDir), jobName)
	case *hpojobv1.HyperparameterTuningJob:
		return fmt.Sprintf("%s/%s/output/model.tar.gz", createOutputPath(prefix, HyperparameterOutputPathSubDir), jobName)
	default:
		return ""
	}
}
