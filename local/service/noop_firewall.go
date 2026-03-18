package service

type NoOpFirewall struct{}

func (f *NoOpFirewall) CreateServerRules(_ string, _, _ []int) error { return nil }
func (f *NoOpFirewall) DeleteServerRules(_ string, _, _ []int) error { return nil }
func (f *NoOpFirewall) UpdateServerRules(_ string, _, _ []int) error { return nil }
