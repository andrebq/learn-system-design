.PHONY: serve-control

serve-control: dist
	./dist/lsd control-plane serve --bind $(LSD_CONTROL_PLANE_BIND)
