// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package vr_xrv

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/nodes"
	"github.com/srl-labs/containerlab/runtime"
	"github.com/srl-labs/containerlab/types"
	"github.com/srl-labs/containerlab/utils"
)

const (
	scrapliPlatformName = "cisco_iosxr"
)

func init() {
	nodes.Register(nodes.NodeKindVrXRV, func() nodes.Node {
		return new(vrXRV)
	})
}

type vrXRV struct {
	cfg     *types.NodeConfig
	mgmt    *types.MgmtNet
	runtime runtime.ContainerRuntime
}

func (s *vrXRV) Init(cfg *types.NodeConfig, opts ...nodes.NodeOption) error {
	s.cfg = cfg
	for _, o := range opts {
		o(s)
	}
	// env vars are used to set launch.py arguments in vrnetlab container
	defEnv := map[string]string{
		"USERNAME":           "clab",
		"PASSWORD":           "clab@123",
		"CONNECTION_MODE":    nodes.VrDefConnMode,
		"DOCKER_NET_V4_ADDR": s.mgmt.IPv4Subnet,
		"DOCKER_NET_V6_ADDR": s.mgmt.IPv6Subnet,
	}
	s.cfg.Env = utils.MergeStringMaps(defEnv, s.cfg.Env)

	if s.cfg.Env["CONNECTION_MODE"] == "macvtap" {
		// mount dev dir to enable macvtap
		s.cfg.Binds = append(s.cfg.Binds, "/dev:/dev")
	}

	s.cfg.Cmd = fmt.Sprintf("--username %s --password %s --hostname %s --connection-mode %s --trace",
		s.cfg.Env["USERNAME"], s.cfg.Env["PASSWORD"], s.cfg.ShortName, s.cfg.Env["CONNECTION_MODE"])

	return nil
}
func (s *vrXRV) Config() *types.NodeConfig { return s.cfg }

func (s *vrXRV) PreDeploy(configName, labCADir, labCARoot string) error {
	utils.CreateDirectory(s.cfg.LabDir, 0777)
	return nil
}

func (s *vrXRV) Deploy(ctx context.Context) error {
	_, err := s.runtime.CreateContainer(ctx, s.cfg)
	return err
}

func (s *vrXRV) PostDeploy(ctx context.Context, ns map[string]nodes.Node) error {
	return nil
}

func (s *vrXRV) GetImages() map[string]string {
	return map[string]string{
		nodes.ImageKey: s.cfg.Image,
	}
}

func (s *vrXRV) WithMgmtNet(mgmt *types.MgmtNet)        { s.mgmt = mgmt }
func (s *vrXRV) WithRuntime(r runtime.ContainerRuntime) { s.runtime = r }
func (s *vrXRV) GetRuntime() runtime.ContainerRuntime   { return s.runtime }

func (s *vrXRV) Delete(ctx context.Context) error {
	return s.runtime.DeleteContainer(ctx, s.Config().LongName)
}

func (s *vrXRV) SaveConfig(ctx context.Context) error {
	err := utils.SaveCfgViaNetconf(s.cfg.LongName,
		nodes.DefaultCredentials[s.cfg.Kind][0],
		nodes.DefaultCredentials[s.cfg.Kind][1],
		scrapliPlatformName,
	)

	if err != nil {
		return err
	}

	log.Infof("saved %s running configuration to startup configuration file\n", s.cfg.ShortName)
	return nil
}
