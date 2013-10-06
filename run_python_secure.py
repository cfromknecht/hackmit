import sys
from sandbox import Sandbox, SandboxConfig
from os import listdir
from StringIO import StringIO

if __name__ == "main":
	'''
	Will run a given code in sandbox (stream stdio from cmd)
	'''
	if len(sys.argv) != 2:
		print "Not enough args"
		sys.exit()
	args = sys.argv
	code = args[1]
	try:
		sandbox = Sandbox(SandboxConfig('stdin', 'stdout'))
		sandbox.io(sys.stdin, sys.stdout)
		sandbox.execute(code)
	except Exception, e:
		print e
