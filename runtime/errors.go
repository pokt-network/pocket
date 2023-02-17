package runtime

import (
	"fmt"
)

const ModuleNotRegisteredError = "module %s not found, did you call ModulesRegistry.RegisterModule() ?"

func ErrModuleNotRegistered(moduleName string) error {
	return fmt.Errorf(ModuleNotRegisteredError, moduleName)
}
