import sys
from sandbox import Sandbox, SandboxConfig
from os import listdir
from StringIO import StringIO

if __name__ == "main":
	if len(sys.argv) != 3:
		print "Not enough args"
		sys.exit()
	args = sys.argv
	problem = '1'
	code = '''print 2'''
	counter = 1
	sandbox = Sandbox(SandboxConfig('stdin', 'stdout'))
	for x in listdir('./question/' + problem ):
		if x[-3:] == ".in":
			print "Attempting: ", x
			f = open(x)
			fIn = StringIO(f.read())
			f.close()
			fOut = StringIO()
			old_stdout = sys.stdout
			old_stdin = sys.stdin
			sys.stdout = fOut
			sys.stdin = fIn
			status = "NOT PROCESSED"
			try:
				sandbox.execute(code)
				status = "PASS"
			except Exception, e:
				pass
			sys.stdin = old_stdin
			sys.stdout = old_stdout
			result = fOut.getstring()
			f = open(x[:-4] + ".ans")
			excected = f.read()
			f.close()
			if result != excected:
				status = "FAIL"
			print "Test case " + str(counter) + ": " + status
			if status != "PASS":
				print fIn.getstring()
			counter = counter + 1
