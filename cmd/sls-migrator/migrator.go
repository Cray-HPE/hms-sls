// MIT License
//
// (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"fmt"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"

	"github.com/Cray-HPE/hms-xname/xnametypes"
	"go.uber.org/zap"
)

type SLSClientInterface interface {
	GetAllHardware(ctx context.Context) ([]sls_common.GenericHardware, error)
	PutHardware(ctx context.Context, hardware sls_common.GenericHardware) error
}

type Migrator struct {
	logger         *zap.Logger
	slsClient      SLSClientInterface
	performChanges bool
}

func (m *Migrator) Run(ctx context.Context) error {
	m.logger.Info("Beginning migration")

	// Retrieve Hardware
	allHardware, err := m.slsClient.GetAllHardware(ctx)
	if err != nil {
		return err
	}

	// For each hardware object
	// - Check for presence of malformed Chassis data
	// - Check for malformed derived fields: Parent/Type/TypeString
	for _, hardware := range allHardware {
		if err := m.migrateMalformedChassisData(ctx, hardware); err != nil {
			m.logger.With(zap.Error(err), zap.Any("hardware", hardware)).Error("Encountered error while attempting to migrate malformed chassis data for hardware object")
		}
		if err := m.migrateMalformedDerivedFields(ctx, hardware); err != nil {
			m.logger.With(zap.Error(err), zap.Any("hardware", hardware)).Error("Encountered error while attempting to migrate malformed derived fields hardware object")
		}
	}

	m.logger.Info("Migration complete")
	return nil
}

func (m *Migrator) migrateMalformedChassisData(ctx context.Context, hardware sls_common.GenericHardware) error {
	expectedTypeString := xnametypes.GetHMSType(hardware.Xname)

	// Look for Chassis that have incorrect Type/TypeString information.
	// This was from bad config generation in CSI in CSM 1.0 and before.
	if (hardware.Class == sls_common.ClassHill || hardware.Class == sls_common.ClassMountain) &&
		hardware.TypeString == xnametypes.ChassisBMC && expectedTypeString == xnametypes.Chassis {

		// Now lets create a ChassisBMC object for the Chassis. Every liquid-cooled chassis gets a ChassisBMC
		// The Chassis will be correct with the normalization below.
		chassisBMC := sls_common.NewGenericHardware(fmt.Sprintf("%sb0", hardware.Xname), hardware.Class, nil)

		m.logger.Debug("Found malformed Chassis object, attempting to correct",
			zap.Any("hardware", hardware),
			zap.Any("chassisBMC", chassisBMC),
		)

		if m.performChanges {
			if err := m.slsClient.PutHardware(ctx, chassisBMC); err != nil {
				return err
			}
		}

		m.logger.Info("Corrected malformed Chassis by creating an accompanying ChassisBMC",
			zap.String("chassis", hardware.Xname), zap.String("chassisBMC", chassisBMC.Xname),
		)
	}

	return nil
}

func (m *Migrator) migrateMalformedDerivedFields(ctx context.Context, hardware sls_common.GenericHardware) error {
	// If Parent/Type/String have unexpected values, perform a PUT to SLS.
	// When the object is PUT back into SLS the derived fields will be recalculated.

	expectedParent := xnametypes.GetHMSCompParent(hardware.Xname)
	expectedTypeString := xnametypes.GetHMSType(hardware.Xname)
	expectedType := sls_common.HMSTypeToHMSStringType(hardware.TypeString)

	if hardware.Parent != expectedParent || hardware.TypeString != expectedTypeString || hardware.Type != expectedType {
		m.logger.Debug("Found malformed hardware with inconsistent derived fields, attempting to correct",
			zap.Any("hardware", hardware))

		if m.performChanges {
			if err := m.slsClient.PutHardware(ctx, hardware); err != nil {
				return err
			}
		}

		m.logger.Info("Corrected hardware with inconsistent derived fields", zap.String("xname", hardware.Xname))
	}

	return nil
}
