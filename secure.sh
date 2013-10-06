PROBLEM="1"
$string
for i in $(ls question/1/*.in)
do
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer.=".ans"
	# echo $base
	# diff <(python run_python_secure.py '''print 2'''<< $base".in") <($base".ans")
	python run_python_secure.py "print 2" < $i
done
