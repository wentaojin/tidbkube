#!/bin/bash

date=`date +'%Y-%m-%d_%H%M%S'`

### BechmarkSQL run dir,Can not with "/"
benchmarkDir=/data/tidb/wentaojin/benchmarksql/run
### BenchmarkSQL test script
benchmarkScript=/data/tidb/wentaojin/benchmarksql/run/runBenchmark.sh
### BechmarkSQL mysql configuration file  name
benchmarkConfName=props.mysql

### Get benchmark runtime(Mins)
runStr=`grep 'runMins' ${benchmarkDir}/${benchmarkConfName}`
runTime=${runStr##*=}

### BechmarkSQL log
benchmarkLog=${benchmarkDir}/benchmark_${date}.log
if [ ! -f ${benchmarkLog} ]; then
	touch ${benchmarkLog}
fi

###########################################
#	Nmon Variable	                      |
###########################################	
# Whether to activate nmon, 1 show enabled Or 0 show disabeld
nmonState=1
# Confirm all machine nmon binary exist the same dir,Then set nmon program execute dir,Can not with "/"
nmonDir=/data/tidb/wentaojin
# How many seconds to set up to collect once (10s collect once)
# If you want Nmon run 10 mins,you can set 10s collect once and collect 66 times
timeInterval=10
# Set how many times to collect (collect 66 times)
collectTimes=13
### Set need run nmon process host IP in TiDB node
hostArry=(172.16.30.86 172.16.30.87 172.16.30.88 172.16.30.89)



### Enable Nmon processes

function enableNmonRun(){
        echo "#--------------------------------------------------------------------------------"
        echo "                            Start Enabled Nmon                                   "
        echo "#--------------------------------------------------------------------------------"
        for ip in ${hostArry[*]}
        do
                # View remote machine whether existed nmon log dir and  nmon binary whether exist
		if `ssh ${ip} test -d ${nmonDir}`; then
			if `ssh ${ip} test -x ${nmonDir}/nmon`; then
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				fi
			else
				ssh ${ip} "chmod +x ${nmonDir}/nmon"
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				fi
			fi
                else
			ssh ${ip} "mkdir -p ${nmonDir}"
			if `ssh ${ip} test -x ${nmonDir}/nmon`; then
                        	if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
                               		ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
                       		else
                                	ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
                        	fi
			else
				scp ${nmonDir}/nmon ${ip}:${nmonDir}/
				ssh ${ip} "chmod +x ${nmonDir}/nmon"
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
				fi
			fi

                fi
        done
	
	echo ""
	if [ $? -eq 0 ]; then
		echo "Enable nmon success."
	else
		echo "Enable nmon failed."

	fi

}



###	Start benchmarkSQL Test
function startBenchmarkSQLTest(){
	
	nohup ${benchmarkScript} ${benchmarkDir}/${benchmarkConfName} > ${benchmarkLog} 2>&1 &
}

### Get Benchmarksql result
function getBenchmarkResult(){

	echo "#--------------------------------------------------------------------------------"
        echo "                     Start Gather Benchmark Results                              "
        echo "#--------------------------------------------------------------------------------"
	

	### Need to view the parameter  values in the props.mysql configuration file
	resultDir=`ls -ldrt ${benchmarkDir}/my_result_* | tail -1 | awk '{print $9}'`
	
	if [ ${nmonState} -eq 1 ]; then
		nmonDir=`ls -ldrt ${benchmarkDir}/nmonLog/* | tail -1 | awk '{print $9}'`	
	fi
	
	rdir=${resultDir##*/}
	echo ""
	echo "Start display Original dir ${rdir} results."
	echo "#<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<" 
	echo ""
	echo "benchmark result name: ${resultDir}"

	echo ""
	echo "benchmark log output name: ${benchmarkLog}"		

	if [ ${nmonState} -eq 1 ]; then
		echo ""
		echo "nmon log dir name: ${nmonDir}"
	fi

	echo ""
	c=`cat ${resultDir}/data/result.csv | awk -F "," '{print $1}' | uniq | tail -1`
	echo "Concurrency: ${c}"
				
	echo ""
	cat ${resultDir}/data/result.csv | awk -F "," 'BEGIN{sum=0}{sum+=$4}END{print "dblatencySum: "sum; print "dblatencyAvg: " sum/NR}'
				
	echo ""
	cat ${resultDir}/data/result.csv | awk -F "," 'BEGIN{sum=0}{sum+=$3}END{print "latencySum: "sum; print "latencyAvg: " sum/NR}'

	echo ""
	nt=`grep 'Measured tpmC (NewOrders)' ${benchmarkLog} | awk '{print $11}'`
	echo "Measured tpmC (NewOrders): ${nt}"		
		
	echo ""
	nl=`grep 'Measured tpmTOTAL ' ${benchmarkLog} | awk '{print $10}'`
	echo "Measured tpmTOTAL: ${nl}"
				
	echo ""
	nst=`grep 'Session Start' ${benchmarkLog} | awk '{print $10" "$11}'`
	echo "tpc-c start run time: ${nst}"
				
	echo ""
	nse=`grep 'Session End' ${benchmarkLog} | awk '{print $10" "$11}'`
	echo "tpc-c end run time: ${nse}"
                
	echo ""
	tc=`grep 'Transaction Count' ${benchmarkLog} | awk '{print $10}'`
	echo "Transaction Count: ${tc}"
	echo "#<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"	
		
	echo ""
		

	echo "Start redirecting the above sort out results to ${benchmarkDir}/results/result.result"	
	concurrency=`cat ${resultDir}/data/result.csv | awk -F "," '{print $1}' | uniq | tail -1`
	dblatency=`cat ${resultDir}/data/result.csv | awk -F "," 'BEGIN{sum=0}{sum+=$4}END{print "dblatencyAvg: " sum/NR}'`
	latency=`cat ${resultDir}/data/result.csv | awk -F "," 'BEGIN{sum=0}{sum+=$3}END{print "latencyAvg: " sum/NR}'`
	tpmC=`grep 'Measured tpmC (NewOrders)' ${benchmarkLog} | awk '{print $11}'`
	tpmcTotal=`grep 'Measured tpmTOTAL ' ${benchmarkLog} | awk '{print $10}'`
	startTime=`grep 'Session Start' ${benchmarkLog} | awk '{print $10" "$11}'`
	endTime=`grep 'Session End' ${benchmarkLog} | awk '{print $10" "$11}'`
	tcounts=`grep 'Transaction Count' ${benchmarkLog} | awk '{print $10}'`

	if [ ! -d "${benchmarkDir}/results" ]; then
		mkdir -p ${benchmarkDir}/results
	fi

		
	echo "benchmarkResult dir name: ${resultDir}" >> ${benchmarkDir}/results/result.result
	echo "benchmarkLog name: ${benchmarkLog}" >> ${benchmarkDir}/results/result.result
	
        if [ ${nmonState} -eq 1 ]; then
                echo "nmonLog dir name: ${nmonDir}" >> ${benchmarkDir}/results/result.result
        fi

	echo "concurrency: ${concurrency}"   >> ${benchmarkDir}/results/result.result
	echo "${dblatency}"		     >> ${benchmarkDir}/results/result.result
	echo "${latency}"		     >> ${benchmarkDir}/results/result.result
	echo "tpmC: ${tpmC}"                 >> ${benchmarkDir}/results/result.result
	echo "tpmcTotal: ${tpmcTotal}"       >> ${benchmarkDir}/results/result.result
	echo "startTime: ${startTime}"       >> ${benchmarkDir}/results/result.result
	echo "endTime: ${endTime}"           >> ${benchmarkDir}/results/result.result
	echo "transaction counts: ${tcounts}">> ${benchmarkDir}/results/result.result
	echo ""				     >> ${benchmarkDir}/results/result.result		

	b=''
	i=0
	while [ $i -le 100 ]
	do
		printf "[%-50s] %d%% \r" "$b" "$i";
		sleep 0.02
		((i=i+2))
		b+='#'
	done
	echo
	echo "Finished redirecting the above sort out results to ${benchmarkDir}/results/result.result"

	echo ""
		
	echo "#<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	echo "Please check result output,Or manual view ${benchmarkDir}/results/result.result"
	echo ""
	cat ${benchmarkDir}/results/result.result
}

### Scp Nmon log to benchmark dir 
function scpNmonLog(){
	echo "#--------------------------------------------------------------------------------"
	echo "       Nmon Log Start Scp From All TiDB Nodes To Benchmark Node Dir              "
	echo "#--------------------------------------------------------------------------------"

	if [ ! -d "${benchmarkDir}/nmonLog" ]; then
		mkdir -p ${benchmarkDir}/nmonLog/
	fi

	if [ ! -d "${benchmarkDir}/nmonLog/${date}" ]; then
		mkdir -p ${benchmarkDir}/nmonLog/${date}
	fi

	for ip in ${hostArry[*]}
   	do
		scp -r ${ip}:${nmonDir}/nmonLog/${date}/* ${benchmarkDir}/nmonLog/${date}/
	done
}


### Stop benchmark Test and Nmon 
function stopBenchmarkTest(){

	# kill Benchmarksql processes in the benchmarksql node
        echo "#--------------------------------------------------------------------------------"
        echo "                         Start Kill runBenchmark Processes                       "
        echo "#--------------------------------------------------------------------------------"

	echo ""
	ps -ef | grep runBenchmark.sh | grep -v grep | awk '{print $2}' | xargs sudo kill -9
	ps -ef | grep runBenchmark.sh | grep -v grep
	if [ $? -eq 0 ]; then
		echo "Localhost processes runBechmark.sh has killed Failed,Please check and manual kill."
	else
		echo "Localhost processes runBechmark.sh has killed Success."
		
	fi
	
	# batch kill nmon processes in all tidb nodes
        echo "#--------------------------------------------------------------------------------"
        echo "                         Start Kill Nmon Processes                               "
        echo "#--------------------------------------------------------------------------------"

	echo ""
	for ip in ${hostArry[*]}
	do
		ssh ${ip} "ps -ef | grep nmon | grep -v grep | awk '{print $2}' | xargs sudo kill -9"
		result=`ssh ${ip} "ps -ef | grep nmon | grep -v grep"`
		if [ ! "$result" = "" ]; then
			echo "TiDB node ${ip} processes Nmon has killed Failed,Please check and manual kill."
		else
			echo "TiDB node ${ip} processes Nmon has killed Success."
		
		fi
	done

        # kill benchmarkSQL processes in the benchamrsql node
        echo "#--------------------------------------------------------------------------------"
        echo "                         Start Kill benchmarkSQL Processes                       "
        echo "#--------------------------------------------------------------------------------"
        echo ""
        ps -ef | grep benchmarkSQL.sh | grep -v grep | awk '{print $2}' | xargs sudo kill -9
        ps -ef | grep benchmarkSQL.sh | grep -v grep
        if [ $? -eq 0 ]; then
                echo "Localhost processes bechmarkSQL.sh has killed Failed,Please check and manual kill."
        else
                echo "Localhost processes bechmarkSQL.sh has killed Success."

        fi
}


### Query Benchmarksql and Nmon status
function statusBenchmarkTest(){
	# query nodes benchmarkSQL processes
        echo "#--------------------------------------------------------------------------------"
        echo "                  Localhost Node BenchmarkSQL Processes                          "
        echo "#--------------------------------------------------------------------------------"
	echo ""

	ps -ef | grep runBenchmark.sh | grep -v grep
	if [ $? -eq 0 ]; then
		echo "Localhost processes runBechmark.sh Existed and Running."
	else
		echo "Localhost processes runBechmark.sh Not existed,Please check and manual query."
	fi
	
	# query all nodes nmon processes

	echo "#--------------------------------------------------------------------------------"
        echo "                       All TiDB Nodes Nmon Processes                             "
        echo "#--------------------------------------------------------------------------------"
	echo ""
	for ip in ${hostArry[*]}
	do	
		result=`ssh ${ip} "ps -ef | grep nmon | grep -v grep"`
		if [ ! "$result" = "" ]; then
			echo ${result}
			echo "TiDB node ${ip} processes Nmon Existed and Running."
			echo ""
		else
			if [ ${nmonState} -eq 0 ];then
				echo "Because variable nmonState equal 0,Show nmon need not turn on,That State is normal state."
			else
				echo "TiDB node ${ip} processes Nmon Not existed,Please manually check why nmon not be turned on,Or Manually start all tidb node nmon processed."
			fi
		fi
	done
}


### Final run benchmark Test
function runBenchmarkTest(){
	if [ ${nmonState} -eq 1 ]; then
		enableNmonRun
		startBenchmarkSQLTest
		statusBenchmarkTest
		

		# check all tidb node nmon process if existed
		declare -a nmonArry
		index=0
		while true
		do
			for ip in ${hostArry[*]}
			do
				result=`ssh ${ip} "ps -ef | grep nmon | grep -v grep"`
				if [ "$result" = "" ]; then
					nmonArry[index]=${ip}
					((index++))
                        	fi		
        		done
		
			nmonarr=($(echo ${nmonArry[*]} |  sed 's/ /\n/g' | sort | uniq))
			num=${#nmonarr[@]}
			hostNums=${#hostArry[@]}
			
                	if [ ${num} -eq ${hostNums} ]; then
				sleep 1
                        	scpNmonLog
				break;
			else
				echo "nmon process existed" > /dev/null 2>&1
			fi
		done

		# check runBenchmark.sh processes if exited
                while true
                do
                        benchmarkPID=`ps -ef | grep runBenchmark.sh | grep -v grep | awk '{print $2}'`
                        if [ "$benchmarkPID" = "" ]; then
                                sleep 1
                                getBenchmarkResult
                                break;
                        else
                                echo "runBenchmark.sh process exsits" > /dev/null 2>&1
                        fi
                done
		
		
	else
		startBenchmarkSQLTest
		statusBenchmarkTest
		
		# check runBenchmark.sh processes if exited
		while true
		do
 			benchmarkPID=`ps -ef | grep runBenchmark.sh | grep -v grep | awk '{print $2}'` 
        		if [ "$benchmarkPID" = "" ]; then
                		sleep 1
                		getBenchmarkResult
				break;
        		else
               			echo "runBenchmark.sh process exsits" > /dev/null 2>&1
        		fi
		done	
	
	fi
}

### Main program entry
if [ ! -n "$1" ]; then
	echo ""
	echo "Parameter can not be Null,Please input parameter start Or stop Or status Or getresult."
	echo ""
else
	case "$1" in
		(start)
			runBenchmarkTest
		;;
		(stop)
			stopBenchmarkTest
		;;

		(getresult)
			getBenchmarkResult
		;;
		(status)
			statusBenchmarkTest
		;;
	esac
fi