PROBLEM="1"
$string
for i in $(ls question/1/*.in)
do
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer=$answer".ans"
	echo python run_python_secure.py "x = raw_input()\nprint x" << $i
	# diff <(python run_python_secure.py "x = raw_input()\nprint x" << $i) <(cat $answer)
done
