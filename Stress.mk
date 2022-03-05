.PHONY: start-stress-test serve-stress-test

start-stress-test: dist
	./dist/lsd stress start --stressor http://$(LSD_STRESS_TEST_SERVE)/start-test --target $(target)

# mkdir -p localfiles
# echo "GET http://localhost:$(LOCAL_PORT)/stress" | \
# 	vegeta attack -name=100qps -workers=1000 -rate=10000/s -duration=1m > ./localfiles/results.bin
# cat ./localfiles/results.bin | vegeta plot > ./localfiles/plot.html
# cat ./localfiles/results.bin | vegeta report -type='hist[0,0.5ms,0.7ms,1ms,5ms,10ms,100ms,500ms,1s]'

serve-stress-test: dist
	./dist/lsd stress serve --bind $(LSD_STRESS_TEST_SERVE)
