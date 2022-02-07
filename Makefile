.PHONY: *

mocks:
	for i in `find . -name mock_*`; do rm -f $i; done
	mockery --all --inpackage --case snake
