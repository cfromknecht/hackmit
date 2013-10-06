PROBLEM="1"
CODE="x = raw_input()
print x"
$string
for i in $(ls question/1/*.in)
do
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer=$answer".ans"
	diff <(cat $i | python run_python_secure.py "$CODE") <(cat $answer)
done
