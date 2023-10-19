package workers

import (
	"testing"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/stretchr/testify/require"
)

type TestWorker struct {
	Worker
	PrivateVariable int
}

func TestInterface(t *testing.T) {
	test1 := TestWorker{
		Worker: Worker{
			Account: models.UpworkAccount{
				Email: "Test1",
			},
			IsActive: true,
		},
		PrivateVariable: 1,
	}

	test2 := TestWorker{
		Worker: Worker{
			Account: models.UpworkAccount{
				Email: "Test2",
			},
			IsActive: true,
		},
		PrivateVariable: 2,
	}

	mapTest := make(map[string]IWorker)
	mapTest["test1"] = &test1
	mapTest["test2"] = &test2

	caseToTest := "test2"

	mapTest[caseToTest].Stop()

	require.Equal(t, test1.IsActive, true)
	require.Equal(t, test2.IsActive, false)
}
