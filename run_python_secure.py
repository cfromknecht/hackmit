import sys
from sandbox import Sandbox, SandboxConfig
from os import listdir
from StringIO import StringIO

if __name__ == "main":
	if len(sys.argv) != 3:
		sys.exit("Not enough args")
	args = sys.argv
	problem = args[1]
	# code = args[2]
	code = '''while True:
		pass
	'''
	diff = ""
	counter = 1
	sandbox = Sandbox(SandboxConfig('stdin', 'stdout'))
	for x in listdir('./question/' + problem):
		if x[-4:] == ".ans":
			try:
				fInput = StringIO(open(x).read())
				result = StringIO()
				sys.stdout = result
				sys.stdin = fInput
				sandbox.execute(code)
				sys.stdout = sys.__stdout__
				sys.stdin = sys.__stdin__
				result_string = result.getvalue()
				fInput_string = fInput.getvalue()
				if result_string != fInput_string:
					status = "FAIL"
				else:
					status = "PASS"
			except Exception, e:
				status = str(e)
			finally:
				diff = diff + "Test case " + str(counter) + ": " + status + "\n"
				if status is not "PASS":
					diff = diff + fInput_string + "\n"
				diff = diff + "\n\n"
			counter = counter + 1
	print diff
