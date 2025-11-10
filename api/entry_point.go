package main

import (
	"fmt"
	"pybot-simulator/api/services"
)


func InitializeOrUpdatePeriod() error {
	services.NewBackup().Start()

	r, err := services.NewRegisterPeriods()
	if err != nil {
		return fmt.Errorf("error al inicializar RegisterPeriods: %w", err)
	}

	// first_period = r.statusPeriod()
	firstPeriod, err := r.StatusPeriod()
	if err != nil {
		return fmt.Errorf("error al checar el status del período: %w", err)
	}

	// if first_period:
	if firstPeriod {
		fmt.Println("Hacer el primer periodo")

		// r.createNewPeriod()
		if err := r.CreateNewPeriod(); err != nil {
			return fmt.Errorf("error al crear nuevo período: %w", err)
		}

		// r.createVoidReading()
		if err := r.CreateVoidReading(); err != nil {
			return fmt.Errorf("error al crear lectura vacía: %w", err)
		}

		fmt.Println("Primer período creado exitosamente.")

	} else {
		// else:
		fmt.Println("Proceder a calcular el ultimo periodo")

		// r.completeLastPeriod()
		if err := r.CompleteLastPeriod(); err != nil {
			return fmt.Errorf("error al completar el último período: %w", err)
		}
		
		fmt.Println("Último período completado, nuevo período iniciado.")
	}

	// Si todo salió bien, no devolvemos ningún error
	return nil
}
