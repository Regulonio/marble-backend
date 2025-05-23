package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/checkmarble/marble-backend/mocks"
	"github.com/checkmarble/marble-backend/models"
	"github.com/checkmarble/marble-backend/models/ast"
	"github.com/checkmarble/marble-backend/usecases/ast_eval"
	"github.com/checkmarble/marble-backend/utils"
)

func TestValidateScenarioIterationImpl_Validate(t *testing.T) {
	ctx := utils.StoreLoggerInContext(context.Background(), utils.NewLogger("text"))
	scenario := models.Scenario{
		Id:                uuid.New().String(),
		OrganizationId:    uuid.New().String(),
		Name:              "scenario_name",
		Description:       "description",
		TriggerObjectType: "object_type",
		CreatedAt:         time.Now(),
		LiveVersionID:     utils.Ptr(uuid.New().String()),
	}

	scenarioIterationID := uuid.New().String()
	scenarioIteration := models.ScenarioIteration{
		Id:             scenarioIterationID,
		OrganizationId: scenario.OrganizationId,
		ScenarioId:     scenario.Id,
		Version:        utils.Ptr(1),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TriggerConditionAstExpression: utils.Ptr(ast.Node{
			Constant: true,
		}),
		Rules: []models.Rule{
			{
				Id:                  "rule",
				ScenarioIterationId: scenarioIterationID,
				OrganizationId:      scenario.OrganizationId,
				DisplayOrder:        0,
				Name:                "rule",
				Description:         "description",
				FormulaAstExpression: utils.Ptr(ast.Node{
					Function: ast.FUNC_GREATER,
					Constant: nil,
					Children: []ast.Node{
						{
							Constant: 10,
						},
						{
							Constant: 100,
						},
					},
				}),
				ScoreModifier: 10,
				CreatedAt:     time.Now(),
			},
		},
		ScoreReviewThreshold:         utils.Ptr(100),
		ScoreBlockAndReviewThreshold: utils.Ptr(1000),
		ScoreDeclineThreshold:        utils.Ptr(1000),
		Schedule:                     "schedule",
	}

	exec := new(mocks.Executor)
	executorFactory := new(mocks.ExecutorFactory)
	executorFactory.On("NewExecutor").Once().Return(exec)
	mdmr := new(mocks.DataModelRepository)
	mdmr.On("GetDataModel", ctx, exec, scenario.OrganizationId, false).
		Return(models.DataModel{
			Version: "version",
			Tables: map[string]models.Table{
				"object_type": {
					Name: "object_type",
					Fields: map[string]models.Field{
						"id": {
							DataType: models.Int,
						},
					},
					LinksToSingle: nil,
				},
			},
		}, nil)

	validator := AstValidatorImpl{
		DataModelRepository: mdmr,
		AstEvaluationEnvironmentFactory: func(params ast_eval.EvaluationEnvironmentFactoryParams) ast_eval.AstEvaluationEnvironment {
			return ast_eval.NewAstEvaluationEnvironment().WithoutOptimizations()
		},
		ExecutorFactory: executorFactory,
	}

	siValidator := ValidateScenarioIterationImpl{
		AstValidator: &validator,
	}

	result := siValidator.Validate(ctx, models.ScenarioAndIteration{
		Scenario:  scenario,
		Iteration: scenarioIteration,
	})
	assert.Empty(t, ScenarioValidationToError(result))
}

func TestValidateScenarioIterationImpl_Validate_notBool(t *testing.T) {
	ctx := utils.StoreLoggerInContext(context.Background(), utils.NewLogger("text"))
	scenario := models.Scenario{
		Id:                uuid.New().String(),
		OrganizationId:    uuid.New().String(),
		Name:              "scenario_name",
		Description:       "description",
		TriggerObjectType: "object_type",
		CreatedAt:         time.Now(),
		LiveVersionID:     utils.Ptr(uuid.New().String()),
	}

	scenarioIterationID := uuid.New().String()
	scenarioIteration := models.ScenarioIteration{
		Id:             scenarioIterationID,
		OrganizationId: scenario.OrganizationId,
		ScenarioId:     scenario.Id,
		Version:        utils.Ptr(1),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TriggerConditionAstExpression: utils.Ptr(ast.Node{
			Constant: 100, // should be a boolean, resulting in an error
		}),
		Rules: []models.Rule{
			{
				Id:                  "rule",
				ScenarioIterationId: scenarioIterationID,
				OrganizationId:      scenario.OrganizationId,
				DisplayOrder:        0,
				Name:                "rule",
				Description:         "description",
				FormulaAstExpression: utils.Ptr(ast.Node{
					Function: ast.FUNC_GREATER,
					Constant: nil,
					Children: []ast.Node{
						{
							Constant: 10,
						},
						{
							Constant: 100,
						},
					},
				}),
				ScoreModifier: 10,
				CreatedAt:     time.Now(),
			},
		},
		ScoreReviewThreshold:         utils.Ptr(100),
		ScoreBlockAndReviewThreshold: utils.Ptr(1000),
		ScoreDeclineThreshold:        utils.Ptr(1000),
		Schedule:                     "schedule",
	}

	exec := new(mocks.Executor)
	executorFactory := new(mocks.ExecutorFactory)
	executorFactory.On("NewExecutor").Once().Return(exec)
	mdmr := new(mocks.DataModelRepository)
	mdmr.On("GetDataModel", ctx, exec, scenario.OrganizationId, false).
		Return(models.DataModel{
			Version: "version",
			Tables: map[string]models.Table{
				"object_type": {
					Name: "object_type",
					Fields: map[string]models.Field{
						"id": {
							DataType: models.Int,
						},
					},
					LinksToSingle: nil,
				},
			},
		}, nil)

	validator := AstValidatorImpl{
		DataModelRepository: mdmr,
		AstEvaluationEnvironmentFactory: func(params ast_eval.EvaluationEnvironmentFactoryParams) ast_eval.AstEvaluationEnvironment {
			return ast_eval.NewAstEvaluationEnvironment().WithoutOptimizations()
		},
		ExecutorFactory: executorFactory,
	}

	siValidator := ValidateScenarioIterationImpl{
		AstValidator: &validator,
	}

	result := siValidator.Validate(ctx, models.ScenarioAndIteration{
		Scenario:  scenario,
		Iteration: scenarioIteration,
	})
	assert.NotEmpty(t, ScenarioValidationToError(result))
}

func TestValidationShouldBypassCircuitBreaking(t *testing.T) {
	ctx := utils.StoreLoggerInContext(context.Background(), utils.NewLogger("text"))
	scenario := models.Scenario{
		Id:                uuid.New().String(),
		OrganizationId:    uuid.New().String(),
		Name:              "scenario_name",
		Description:       "description",
		TriggerObjectType: "object_type",
		CreatedAt:         time.Now(),
		LiveVersionID:     utils.Ptr(uuid.New().String()),
	}

	scenarioIterationID := uuid.New().String()
	scenarioIteration := models.ScenarioIteration{
		Id:             scenarioIterationID,
		OrganizationId: scenario.OrganizationId,
		ScenarioId:     scenario.Id,
		Version:        utils.Ptr(1),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TriggerConditionAstExpression: utils.Ptr(ast.Node{
			Constant: true,
		}),
		Rules: []models.Rule{
			{
				Id:                  "rule",
				ScenarioIterationId: scenarioIterationID,
				OrganizationId:      scenario.OrganizationId,
				DisplayOrder:        0,
				Name:                "rule",
				Description:         "description",
				FormulaAstExpression: utils.Ptr(ast.Node{
					Function: ast.FUNC_AND,
					Constant: nil,
					Children: []ast.Node{
						{
							Function: ast.FUNC_EQUAL,
							Children: []ast.Node{
								{Constant: 100},
								{Constant: 101},
							},
						},
						{
							Function: ast.FUNC_EQUAL,
							Children: []ast.Node{
								{Constant: 100},
								{Constant: "oplop"},
							},
						},
					},
				}),
				ScoreModifier: 10,
				CreatedAt:     time.Now(),
			},
		},
		ScoreReviewThreshold:         utils.Ptr(100),
		ScoreBlockAndReviewThreshold: utils.Ptr(1000),
		ScoreDeclineThreshold:        utils.Ptr(1000),
		Schedule:                     "schedule",
	}

	exec := new(mocks.Executor)
	executorFactory := new(mocks.ExecutorFactory)
	executorFactory.On("NewExecutor").Once().Return(exec)
	mdmr := new(mocks.DataModelRepository)
	mdmr.On("GetDataModel", ctx, exec, scenario.OrganizationId, false).
		Return(models.DataModel{
			Version: "version",
			Tables: map[string]models.Table{
				"object_type": {
					Name: "object_type",
					Fields: map[string]models.Field{
						"id": {
							DataType: models.Int,
						},
					},
					LinksToSingle: nil,
				},
			},
		}, nil)

	validator := AstValidatorImpl{
		DataModelRepository: mdmr,
		AstEvaluationEnvironmentFactory: func(params ast_eval.EvaluationEnvironmentFactoryParams) ast_eval.AstEvaluationEnvironment {
			return ast_eval.NewAstEvaluationEnvironment().WithoutOptimizations()
		},
		ExecutorFactory: executorFactory,
	}

	siValidator := ValidateScenarioIterationImpl{
		AstValidator: &validator,
	}

	result := siValidator.Validate(ctx, models.ScenarioAndIteration{
		Scenario:  scenario,
		Iteration: scenarioIteration,
	})
	assert.NotEmpty(t, ScenarioValidationToError(result))
}
