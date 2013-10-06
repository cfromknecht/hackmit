import sys
from sandbox import Sandbox, SandboxConfig

if __name__ == "__main__":
	'''
	Will run a given code in sandbox (stream stdio from cmd)
	arg[1] = code given
	to run
	python run_python_secure.py <code>  < inputfile 
	'''
	if len(sys.argv) != 2:
		print "Not enough args"
		sys.exit()
	args = sys.argv
	code = args[1]
	print "Will run: ", code
	try:
		sandbox = Sandbox(SandboxConfig('stdin', 'stdout'))
		sandbox.io(sys.stdin, sys.stdout)
		sandbox.execute(code)
	except Exception, e:
		print e