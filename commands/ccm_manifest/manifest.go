// Copyright (c) 2026-2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package ccm_manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/ccm/manager"
	"github.com/choria-io/ccm/model"
	"github.com/choria-io/ccm/resources/apply"
	"github.com/choria-io/fisk"
	"github.com/sirupsen/logrus"
)

type Command struct {
	Manifest         string `json:"manifest"`
	NATSContext      string `json:"nats_context"`
	RenderSummary    bool   `json:"render_summary"`
	NoRenderMessages bool   `json:"no_render_messages"`

	Transform *builder.Transform `json:"transform"`

	builder.GenericCommand
	builder.GenericSubCommands
}

type CCMManifest struct {
	def       *Command
	arguments map[string]any
	flags     map[string]any
	mgr       model.Manager
	cmd       *fisk.CmdClause
	ctx       context.Context
	log       builder.Logger
	b         *builder.AppBuilder
}

func Register() error {
	return builder.RegisterCommand("exec", NewCCMManifest)
}

func MustRegister() {
	builder.MustRegisterCommand("ccm_manifest", NewCCMManifest)
}

func NewCCMManifest(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	manifest := &CCMManifest{
		b:         b,
		ctx:       b.Context(),
		log:       log,
		def:       &Command{},
		arguments: map[string]any{},
		flags:     map[string]any{},
	}

	err := json.Unmarshal(j, manifest.def)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	return manifest, nil
}

func (r *CCMManifest) String() string { return fmt.Sprintf("%s (ccm_manifest)", r.def.Name) }

func (r *CCMManifest) Validate(log builder.Logger) error {
	if r.def.NATSContext == "" {
		r.def.NATSContext = "CCM"
	}

	if r.def.Transform != nil {
		err := r.def.Transform.Validate(log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CCMManifest) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *CCMManifest) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.arguments, r.flags, r.b, r.runCommand)

	return r.cmd, nil
}

func (r *CCMManifest) runCommand(_ *fisk.ParseContext) error {
	url, err := builder.ParseStateTemplate(r.def.Manifest, r.arguments, r.flags, r.b.Configuration())
	if err != nil {
		return fmt.Errorf("invalid manifest url template: %w", err)
	}

	manifestData := map[string]any{}
	for k, v := range r.arguments {
		manifestData[k] = v
	}
	for k, v := range r.flags {
		manifestData[k] = v
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	log := manager.NewLogrusLogger(logger.WithFields(logrus.Fields{
		"component": "builder",
	}))

	mgr := r.mgr
	if mgr == nil {
		mgr, err = manager.NewManager(log, log, manager.WithNatsContext(r.def.NATSContext))
		if err != nil {
			return err
		}
	}
	mgr.SetData(manifestData)

	_, m, wd, err := apply.ResolveManifestUrl(r.ctx, mgr, url, log, apply.WithOverridingResolvedData(manifestData))
	if err != nil {
		return err
	}

	if wd != "" {
		mgr.SetWorkingDirectory(wd)
		defer os.RemoveAll(wd)
	}

	if m.PreMessage() != "" {
		fmt.Fprintln(r.b.Stdout(), m.PreMessage())
	}

	_, err = m.Execute(r.ctx, mgr, false, log)
	if err != nil {
		return err
	}

	if m.PostMessage() != "" {
		fmt.Fprintln(r.b.Stdout(), m.PostMessage())
	}

	summary, err := mgr.SessionSummary()
	if err != nil {
		return err
	}

	if r.def.RenderSummary {
		fmt.Println()
		summary.RenderText(r.b.Stdout())
	}

	if r.def.Transform != nil {
		sj, err := json.Marshal(summary)
		if err != nil {
			return err
		}

		r.log.Warnf("Summary: %s", string(sj))
		tRes, err := r.def.Transform.TransformBytes(r.ctx, sj, r.arguments, r.flags, r.b)
		if err != nil {
			return err
		}
		fmt.Fprintln(r.b.Stdout(), string(tRes))
	}

	return nil
}
