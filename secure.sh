PROBLEM="1"
$string
for i in $(ls question/1/*.in)
do
	base=$(echo $i | rev | cut -c 4- | rev)
	# echo $base
	diff <(python run_python_secure.py '''print 2'''<< $base".in") <($base".ans")
done
