package project_test

import (
	"testing"

	"github.com/lxc/lxd/lxd/db"
	"github.com/lxc/lxd/lxd/instance/instancetype"
	"github.com/lxc/lxd/lxd/project"
	"github.com/lxc/lxd/shared/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// If there's no limit configured on the project, the check passes.
func TestCheckLimitsUponInstanceCreation_NotConfigured(t *testing.T) {
	tx, cleanup := db.NewTestClusterTx(t)
	defer cleanup()

	req := api.InstancesPost{
		Name: "c1",
		Type: api.InstanceTypeContainer,
	}
	err := project.CheckLimitsUponInstanceCreation(tx, "default", req)
	assert.NoError(t, err)
}

// If a limit is configured and the current number of instances is below it, the check passes.
func TestCheckLimitsUponInstanceCreation_Below(t *testing.T) {
	tx, cleanup := db.NewTestClusterTx(t)
	defer cleanup()

	_, err := tx.ProjectCreate(api.ProjectsPost{
		Name: "p1",
		ProjectPut: api.ProjectPut{
			Config: map[string]string{
				"limits.containers": "5",
			},
		},
	})
	require.NoError(t, err)

	_, err = tx.InstanceCreate(db.Instance{
		Project:      "p1",
		Name:         "c1",
		Type:         instancetype.Container,
		Architecture: 1,
		Node:         "none",
	})
	require.NoError(t, err)

	req := api.InstancesPost{
		Name: "c2",
		Type: api.InstanceTypeContainer,
	}

	err = project.CheckLimitsUponInstanceCreation(tx, "p1", req)
	assert.NoError(t, err)
}

// If a limit is configured and it matches the current number of instances, the
// check fails.
func TestCheckLimitsUponInstanceCreation_Above(t *testing.T) {
	tx, cleanup := db.NewTestClusterTx(t)
	defer cleanup()

	_, err := tx.ProjectCreate(api.ProjectsPost{
		Name: "p1",
		ProjectPut: api.ProjectPut{
			Config: map[string]string{
				"limits.containers": "1",
			},
		},
	})
	require.NoError(t, err)

	_, err = tx.InstanceCreate(db.Instance{
		Project:      "p1",
		Name:         "c1",
		Type:         instancetype.Container,
		Architecture: 1,
		Node:         "none",
	})
	require.NoError(t, err)

	req := api.InstancesPost{
		Name: "c2",
		Type: api.InstanceTypeContainer,
	}

	err = project.CheckLimitsUponInstanceCreation(tx, "p1", req)
	assert.EqualError(t, err, "Reached maximum number of instances of type container in project p1")
}

// If a limit is configured, but for a different instance type, the check
// passes.
func TestCheckLimitsUponInstanceCreation_DifferentType(t *testing.T) {
	tx, cleanup := db.NewTestClusterTx(t)
	defer cleanup()

	_, err := tx.ProjectCreate(api.ProjectsPost{
		Name: "p1",
		ProjectPut: api.ProjectPut{
			Config: map[string]string{
				"limits.containers": "1",
			},
		},
	})
	require.NoError(t, err)

	_, err = tx.InstanceCreate(db.Instance{
		Project:      "p1",
		Name:         "vm1",
		Type:         instancetype.VM,
		Architecture: 1,
		Node:         "none",
	})
	require.NoError(t, err)

	req := api.InstancesPost{
		Name: "c2",
		Type: api.InstanceTypeContainer,
	}

	err = project.CheckLimitsUponInstanceCreation(tx, "p1", req)
	assert.NoError(t, err)
}
