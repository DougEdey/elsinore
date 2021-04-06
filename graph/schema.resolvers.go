package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/dougedey/elsinore/graph/model"
)

func (r *hysteriaSettingsResolver) ID(ctx context.Context, obj *devices.HysteriaSettings) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *manualSettingsResolver) ID(ctx context.Context, obj *devices.ManualSettings) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AssignProbe(ctx context.Context, settings *model.ProbeSettings) (*devices.TemperatureController, error) {
	probe := devices.GetTemperature(*settings.Address)
	if probe != nil {
		return devices.CreateTemperatureController(*settings.Name, probe)
	}
	return nil, fmt.Errorf("Coud not find a probe for %v", *settings.Address)
}

func (r *pidSettingsResolver) ID(ctx context.Context, obj *devices.PidSettings) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Probe(ctx context.Context, address *string) (*devices.TemperatureProbe, error) {
	device := devices.GetTemperature(*address)
	if device != nil {
		return device, nil
	}
	return nil, fmt.Errorf("No device found for address %v", *address)
}

func (r *queryResolver) ProbeList(ctx context.Context) ([]*devices.TemperatureProbe, error) {
	return devices.GetProbes(), nil
}

func (r *queryResolver) FetchProbes(ctx context.Context, addresses []*string) ([]*devices.TemperatureProbe, error) {
	deviceList := []*devices.TemperatureProbe{}
	missingAddresses := []string{}
	for _, address := range addresses {
		device := devices.GetTemperature(*address)
		if device != nil {
			deviceList = append(deviceList, device)
		} else {
			missingAddresses = append(missingAddresses, *address)
		}
	}

	missingError := (error)(nil)
	if len(missingAddresses) > 0 {
		missingError = fmt.Errorf("No device(s) found for address(es): %v", missingAddresses)
	}

	return deviceList, missingError
}

func (r *temperatureControllerResolver) ID(ctx context.Context, obj *devices.TemperatureController) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *temperatureProbeResolver) ID(ctx context.Context, obj *devices.TemperatureProbe) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// HysteriaSettings returns generated.HysteriaSettingsResolver implementation.
func (r *Resolver) HysteriaSettings() generated.HysteriaSettingsResolver {
	return &hysteriaSettingsResolver{r}
}

// ManualSettings returns generated.ManualSettingsResolver implementation.
func (r *Resolver) ManualSettings() generated.ManualSettingsResolver {
	return &manualSettingsResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// PidSettings returns generated.PidSettingsResolver implementation.
func (r *Resolver) PidSettings() generated.PidSettingsResolver { return &pidSettingsResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// TemperatureController returns generated.TemperatureControllerResolver implementation.
func (r *Resolver) TemperatureController() generated.TemperatureControllerResolver {
	return &temperatureControllerResolver{r}
}

// TemperatureProbe returns generated.TemperatureProbeResolver implementation.
func (r *Resolver) TemperatureProbe() generated.TemperatureProbeResolver {
	return &temperatureProbeResolver{r}
}

type hysteriaSettingsResolver struct{ *Resolver }
type manualSettingsResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type pidSettingsResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type temperatureControllerResolver struct{ *Resolver }
type temperatureProbeResolver struct{ *Resolver }
