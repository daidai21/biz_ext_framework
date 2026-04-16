module github.com/daidai21/biz_ext_framework/service_manager

go 1.22

require (
	github.com/daidai21/biz_ext_framework/biz_ctx v0.0.0
	github.com/daidai21/biz_ext_framework/biz_identity v0.0.0
	github.com/daidai21/biz_ext_framework/biz_observation v0.0.0
	github.com/daidai21/biz_ext_framework/biz_process v0.0.0
	github.com/daidai21/biz_ext_framework/ext_interceptor v0.0.0
	github.com/daidai21/biz_ext_framework/ext_model v0.0.0
	github.com/daidai21/biz_ext_framework/ext_process v0.0.0
	github.com/daidai21/biz_ext_framework/ext_spi v0.0.0
)

replace (
	github.com/daidai21/biz_ext_framework/biz_ctx => ../biz_ctx
	github.com/daidai21/biz_ext_framework/biz_identity => ../biz_identity
	github.com/daidai21/biz_ext_framework/biz_observation => ../biz_observation
	github.com/daidai21/biz_ext_framework/biz_process => ../biz_process
	github.com/daidai21/biz_ext_framework/ext_interceptor => ../ext_interceptor
	github.com/daidai21/biz_ext_framework/ext_model => ../ext_model
	github.com/daidai21/biz_ext_framework/ext_process => ../ext_process
	github.com/daidai21/biz_ext_framework/ext_spi => ../ext_spi
)
