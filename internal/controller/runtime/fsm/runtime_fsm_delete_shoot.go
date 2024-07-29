package fsm

import (
	"context"
	gardener_api "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnDeleteShoot(ctx context.Context, m *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	m.log.Info("delete shoot state")
	if !isGardenerCloudDelConfirmationSet(s.shoot.Annotations) {
		m.log.Info("patching shoot with del-confirmation")
		// workaround for Gardener client
		setObjectFields(s.shoot)
		s.shoot.Annotations = addGardenerCloudDelConfirmation(s.shoot.Annotations)

		err := m.ShootClient.Patch(ctx, s.shoot, client.Apply, &client.PatchOptions{
			FieldManager: "kim",
			Force:        ptrTo(true),
		})

		if err != nil {
			m.log.Error(err, "unable to patch shoot:", s.shoot.Name)
			return requeue()
		}
	}

	if !s.instance.IsStateWithConditionSet(imv1.RuntimeStateTerminating, imv1.ConditionTypeRuntimeProvisioned, imv1.ConditionReasonDeletion) {
		m.log.Info("setting state to in deletion")
		s.instance.UpdateStateDeletion(
			imv1.ConditionTypeRuntimeProvisioned,
			imv1.ConditionReasonDeletion,
			"Unknown",
			"Runtime deletion initialised",
		)
		return updateStatusAndRequeue()
	}

	m.log.Info("deleting shoot")
	err := m.ShootClient.Delete(ctx, s.shoot)
	if err != nil {
		m.log.Error(err, "Failed to delete gardener Shoot")
		attemptForceDeletion(m, ctx, s)

		s.instance.UpdateStateDeletion(
			imv1.ConditionTypeRuntimeProvisioned,
			imv1.ConditionReasonGardenerError,
			"False",
			"Gardener API delete error",
		)
		return updateStatusAndRequeueAfter(gardenerRequeueDuration)
	}

	return updateStatusAndRequeueAfter(gardenerRequeueDuration)
}

func attemptForceDeletion(m *fsm, ctx context.Context, s *systemState) {
	shoot := &gardener_api.Shoot{}
	shootKey := types.NamespacedName{
		Name:      s.shoot.Name,
		Namespace: s.shoot.Namespace,
	}
	err := m.Client.Get(ctx, shootKey, shoot)

	supportedErrorCodes := []gardener_api.ErrorCode{
		gardener_api.ErrorCleanupClusterResources,
		gardener_api.ErrorConfigurationProblem,
		gardener_api.ErrorInfraDependencies,
		gardener_api.ErrorInfraUnauthenticated,
		gardener_api.ErrorInfraUnauthorized,
	}

	for _, condition := range shoot.Status.Conditions {
		for _, conditionErrorCode := range condition.Codes {
			for _, supportedErrorCode := range supportedErrorCodes {
				if conditionErrorCode == supportedErrorCode {
					m.log.Info("Force deletion enabled for shoot", "shoot", shoot.Name, "supportedErrorCode", supportedErrorCode)

					// TODO: add annotation for force deletion
					//s.shoot.Annotations = addGardenerCloudDelConfirmation(s.shoot.Annotations)
					err := m.ShootClient.Patch(ctx, s.shoot, client.Apply, &client.PatchOptions{
						FieldManager: "kim",
						Force:        ptrTo(true),
					})
					if err != nil {
						m.log.Error(err, "unable to patch shoot:", s.shoot.Name)
					}
					return
				}
			}
		}
	}
}

func conditionShouldEnableForceDelation()

func isGardenerCloudDelConfirmationSet(a map[string]string) bool {
	if len(a) == 0 {
		return false
	}
	val, found := a[imv1.AnnotationGardenerCloudDelConfirmation]
	return found && (val == "true")
}

func addGardenerCloudDelConfirmation(a map[string]string) map[string]string {
	if len(a) == 0 {
		a = map[string]string{}
	}
	a[imv1.AnnotationGardenerCloudDelConfirmation] = "true"
	return a
}
