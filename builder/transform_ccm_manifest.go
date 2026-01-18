// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/choria-io/ccm/manager"
	"github.com/choria-io/ccm/resources/apply"
)

type ccmManifestTransform struct {
	Manifest         string `json:"manifest"`
	NATSContext      string `json:"nats_context"`
	RenderSummary    bool   `json:"render_summary"`
	NoRenderMessages bool   `json:"no_render_messages"`
}

func newCCMManifestTransform(trans *Transform) (*ccmManifestTransform, error) {
	// copy it
	ccm := *trans.CCMManifest
	if ccm.NATSContext == "" {
		ccm.NATSContext = "CCM"
	}

	return &ccm, nil
}

func (manifest *ccmManifestTransform) Validate(_ Logger) error {
	if manifest.Manifest == "" {
		return fmt.Errorf("a manifest url is required")
	}

	return nil
}

func (manifest *ccmManifestTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	url, err := ParseStateTemplate(manifest.Manifest, args, flags, b.Configuration())
	if err != nil {
		return nil, fmt.Errorf("invalid manifest url template: %v", err)
	}

	manifestData := map[string]any{}
	for k, v := range args {
		manifestData[k] = v
	}
	for k, v := range flags {
		manifestData[k] = v
	}
	err = json.Unmarshal(input, &manifestData)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	log := manager.NewLogrusLogger(logger.WithFields(logrus.Fields{
		"component": "builder",
	}))

	mgr, err := manager.NewManager(log, log, manager.WithNatsContext(manifest.NATSContext))
	if err != nil {
		return nil, err
	}
	mgr.SetData(manifestData)

	_, m, wd, err := apply.ResolveManifestUrl(ctx, mgr, url, log, apply.WithOverridingResolvedData(manifestData))
	if err != nil {
		return nil, err
	}

	if wd != "" {
		mgr.SetWorkingDirectory(wd)
		defer os.RemoveAll(wd)
	}

	if m.PreMessage() != "" {
		fmt.Println(m.PreMessage())
	}

	_, err = m.Execute(ctx, mgr, false, log)
	if err != nil {
		return nil, err
	}

	if m.PostMessage() != "" {
		fmt.Println(m.PostMessage())
	}

	summary, err := mgr.SessionSummary()
	if err != nil {
		return nil, err
	}

	if manifest.RenderSummary {
		fmt.Println()
		summary.RenderText(os.Stdout)

		return bytes.NewReader(nil), nil
	}

	j, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(j), nil
}
