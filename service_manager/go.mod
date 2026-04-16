module github.com/daidai21/biz_ext_framework/service_manager

go 1.22

require (
	github.com/daidai21/biz_ext_framework/biz_identity v0.0.0
	github.com/daidai21/biz_ext_framework/biz_process v0.0.0
	github.com/daidai21/biz_ext_framework/ext_model v0.0.0
)

replace github.com/daidai21/biz_ext_framework/biz_identity => ../biz_identity

replace github.com/daidai21/biz_ext_framework/biz_process => ../biz_process

replace github.com/daidai21/biz_ext_framework/ext_model => ../ext_model
