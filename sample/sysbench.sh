#!/bin/bash

#########################################################################################################################################
#	1、nmon and sysbech run using tidb user,So that scp and kill processes.                             				|
#	2、If you activate nmon, make sure that there is an nmon program on the machine or on the machine with tidb user mutual trust,	|
#	and the tidb user nmon path is accessible.											|
#	3、The sysbench test is best run on a centrally controlled machine or independently with a tidb user mutual trust machine,	|
#	Because all sysbench test runs centrally on one server                                              				|
#########################################################################################################################################
#使用简概要
# 1、直接用tidb用户 放在中控机器上或者带有tidb用户互信得
# 2、是否执行nmon可选以及是否只haproxy 压测或者压测多个tidb同时都可以，POC结果会统一到 sysbench执行机器上 sysbenchLog、nmonLog、results目录
# 3、压测所有tidb节点起sysbench，sysbench进程都在同一个机器上，不是去到每台tidb节点起得
# 4、status 查看状态，stop kill 所有、startk开启压测

date=`date +'%Y%m%d_%H%M%S'`

echo ""
echo -e "\033[43;35m Sysbench Test Start At $date \033[0m \n"

###########################################
#	Host and Database Variable	  |
###########################################	
# Set host IP array
hostArry=(172.16.30.86 172.16.30.89)
# Set host port array,This parameter can  set multiple port,So show one machine install multiple tidb node.
portArry=(5000)
user=root
dbname=sbtest

###########################################
#	Sysbench Variable	          |
###########################################	
# Set sysbench store dir without '/'
sysbenchDir=/data/tidb/wentaojin
# Set sysbench log dir
sysbenchLog=${sysbenchDir}/sysbenchLog/${date}
if [ ! -d ${sysbenchLog} ]; then
	mkdir -p ${sysbenchDir}/sysbenchLog/${date}/
fi



# Set sysbench concurrency
threads=50
# Set sysbench test runtime(s)
runtime=120
# Set sysbench mode,For example: oltp_read_write(oltp),oltp_write_only(insert),oltp_point_select(select)
mode=oltp_read_write
# Set table numbers
tabCounts=32
# Set table size
tabSize=10000000

###########################################
#	Nmon Variable	                  |
###########################################	
# Whether to activate nmon, 1 show enabled Or 0 show disabeld
nmonState=1
# Confirm all machine nmon binary exist the same dir,Then set nmon program execute dir,without '/'
nmonDir=/data/tidb/wentaojin

# How many seconds to set up to collect once (10s collect once), timeInterval and collectTimes value set should see variable sysbench runtime value
# If you want Nmon run 10 mins,you can set 10s collect once and collect 66 times
timeInterval=10
# Set how many times to collect (collect 66 times)
collectTimes=12


### Get the length of array elements
nodeNums=${#portArry[@]}
###	Get the all counts of tidb node
hostNums=${#hostArry[@]}
nodeAlls=$[${hostNums}*${nodeNums}] 


###########################################################################
#	First Part:	Create sysbench TiDB node config                  |
#   function for generate sysbench configuration file                     |
###########################################################################
function genSysbenchConf(){
	echo "#----------------------------------------------------------------------------"
	echo "Existed TiDB nodes in one machine,Beginning create sysbench tidb node config."
	echo "#----------------------------------------------------------------------------"
    	for ip in ${hostArry[*]}
    	do 
		###     Intercept IP field
		ipFiled=`echo ${ip} | awk -F "." '{print $4}'`    
	
		for port in ${portArry[*]}
    		do
			if [ ! -d ${sysbenchDir}/conf ]; then
				mkdir -p ${sysbenchDir}/conf
			fi
			cat << EOF > ${sysbenchDir}/conf/sysbench_${ipFiled}_${port}_config
mysql-host=${ip}
mysql-port=${port}
mysql-user=${user}
mysql-db=${dbname}
time=${runtime}
threads=${threads}
report-interval=10
db-driver=mysql
EOF
    			if [ $? -eq 0 ]; then
    				echo "Create config for TiDB node: ${ip} port: ${port} success."
    			else
    				echo "Create config for TiDB node: ${ip} port: ${port} failed."
    			fi
    		done
    	done	
}



###########################################################################
#       Second Part:    Enable Nmon Run                                   |
#   function for whether enable nmon run                                  |
###########################################################################
function enableNmonRun(){
	for ip in ${hostArry[*]}
        do
                # View remote machine whether existed nmon log dir and  nmon binary whether exist
		if `ssh ${ip} test -d ${nmonDir}`; then
			if `ssh ${ip} test -x ${nmonDir}/nmon`; then
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				fi
			else
				ssh ${ip} "chmod +x ${nmonDir}/nmon"
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				fi
			fi
        else
			ssh ${ip} "mkdir -p ${nmonDir}"
			if `ssh ${ip} test -x ${nmonDir}/nmon`; then
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
                               		ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
                       			ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				else
                                	ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
                        	fi
			else
				scp ${nmonDir}/nmon ${ip}:${nmonDir}/
				ssh ${ip} "chmod +x ${nmonDir}/nmon"
				if `ssh ${ip} test -d  ${nmonDir}/nmonLog/${date}`; then
					ssh ${ip} "${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
				else
					ssh ${ip} "mkdir -p ${nmonDir}/nmonLog/${date};${nmonDir}/nmon -s ${timeInterval} -c ${collectTimes} -f -N -m ${nmonDir}/nmonLog/${date} &"
					ssh ${ip} "ps -ef|grep nmon |grep -v grep"
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


###########################################################################
#       Third Part:    Enable sysbench Run                                 |
#   function for enable sysbench run                                       |
###########################################################################


function RunningSysbench(){
	for ip in ${hostArry[*]}
	do
		###     Intercept IP field
		ipField=`echo ${ip} | awk -F "." '{print $4}'`
		for port in ${portArry[*]}
		do
			nohup sysbench --config-file=${sysbenchDir}/conf/sysbench_${ipField}_${port}_config ${mode} --tables=${tabCounts} --table-size=${tabSize} run > ${sysbenchLog}/sysbench_${ipField}_${port}_${mode}.log 2>&1 &
			result=`ps -ef | grep sysbench | grep -v grep | grep -v sysbench.sh`
			if [ ! "${result}" = "" ]; then
				echo "Start TiDB: ${ip} port: ${port} Success,Please wait."
				sleep 1
			else
				echo ""
				echo "Found Error<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
				echo "Start TiDB: ${ip} port: ${port} Failed,Please Check.                                   "
				echo "Start kill all tidb sysbench test processed,Then manual rerunning.                     "
				stopSysbenchTest
				exit 1
			fi
		done
	done
}


### Check sysbench log if exited fatal

function CheckSysbenchLog(){
        n=0
        times=`expr $((${runtime}/10))`
        while(($n<=${times}))
        do
                fatal=`grep 'FATAL' ${sysbenchLog}/* |grep -v grep| awk -F ":" '{print $1}' | uniq`
                if [ ! "${fatal}" = "" ]; then
                        for i in ${fatal}
                        do
				echo ""
                                echo -e "\033[43;35m Existed sysbench node test error,Please check sysbench log: $i \033[0m \n"
                                echo -e "\033[43;35m Start kill all tidb sysbench test processed,Then manual rerunning. \033[0m \n"
				echo ""
                                stopSysbenchTest
                        done
                fi
                n=$((n + 1))
                sleep 10
        done
}


###########################################################################
#	Fourth Part:	Run TiDB sysbench test                            |
#   function for run TiDB sysbench test                                   |
###########################################################################
function runTiDBTest(){
	# Set nmon log dir in the sysbench node,So that the nmon log of all tidb nodes will be uniformly stored here.
	if [ ${nmonState} -eq 1 ]; then
    		echo "#-----------------------------------------------------------------------------"
    		echo "Nmon program has enabled,All tidb machine is beginning running with user tidb."
    		echo "#-----------------------------------------------------------------------------"
		echo ""

		enableNmonRun
	
 		echo ""
     		echo "#-----------------------------------------------------------------------------"
    		echo "Nmon program has running,Sysbench program is beginning test.                  "
    		echo "#-----------------------------------------------------------------------------"       	
     		echo ""

		RunningSysbench
		CheckSysbenchLog	
		
	else	echo "#--------------------------------------------------------------------------------"
    		echo "Nmon program has disabeld,Sysbench program is beginning test in all tidb machine."
    		echo "#--------------------------------------------------------------------------------"
    		echo ""
		
		RunningSysbench
		CheckSysbenchLog
	fi
}



###########################################################################
#	Five Part:	Get sysbench test results                          |
#   function for get sysbench test results                                 |
###########################################################################

function getSysbenchTestResults(){

	logNums=`ls ${sysbenchLog}/*.log|wc -l`

	### Judge log nums,If log nums equal 1,directly output results,not get avg result
	if [ ${logNums} -ge 2 ]; then

		### Get threads
		thrds=`grep thds ${sysbenchLog}/* | head -1 | awk '{print $5}' | sort -u | uniq`
	
		num=`grep 'transactions:' ${sysbenchLog}/* | wc -l`
		if [ ${num} -ne ${nodeAlls} ]; then
			echo "#-----------------------------------------------------------"
			echo "The sysbench thread exist wrong or not running,Please check."
			echo "#-----------------------------------------------------------"
			exit 1
		fi
	
		###	Get tps
		tps=`grep 'transactions:' ${sysbenchLog}/* | sed 's/(//g' | awk 'BEGIN{s=0}{s+=$4}END{print s/NR}'`
	
		### Get qps
		qps=`grep 'queries:' ${sysbenchLog}/* | sed 's/(//g' | awk 'BEGIN{a=0}{a+=$4}END{print a/NR}'`
	
		###	Get latency
		# 95th
		th=`grep '95th percentile: ' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$4}END{print a/NR}'`
		# min
		min=`grep 'min:' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$3}END{print a/NR}'`
		# avg
		avg=`grep 'avg:' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$3}END{print a/NR}'`
	
		echo ""
		totalCon=`expr $((${thrds}*${num}))`
        	echo "All tidb node sysbench test avg result as follows,Sysbench detailed result see dir ${sysbenchLog}"
        	echo ""
        	echo "Total concurrency ${totalCon} results:        "
        	echo "----------------------------------------------"

        	echo "qps       :       ${qps}"
        	echo "tps       :       ${tps}"
        	echo "latency:          0.95:   ${th}"
        	echo "                  min:    ${min}"
        	echo "                  avg:    ${avg}"

	else
		### Get threads
                thrds=`grep thds ${sysbenchLog}/* | head -1 | awk '{print $5}' | sort -u | uniq`

                num=`grep 'transactions:' ${sysbenchLog}/* | wc -l`
                if [ ${num} -ne ${nodeAlls} ]; then
                        echo "#-----------------------------------------------------------"
                        echo "The sysbench thread exist wrong or not running,Please check."
                        echo "#-----------------------------------------------------------"
                        exit 1
                fi
		###     Get tps
                tps=`grep 'transactions:' ${sysbenchLog}/* | sed 's/(//g' | awk 'BEGIN{s=0}{s+=$3}END{print s/NR}'`

                ### Get qps
                qps=`grep 'queries:' ${sysbenchLog}/* | sed 's/(//g' | awk 'BEGIN{a=0}{a+=$3}END{print a/NR}'`

                ###     Get latency
                # 95th
                th=`grep '95th percentile: ' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$3}END{print a/NR}'`
                # min
                min=`grep 'min:' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$2}END{print a/NR}'`
                # avg
                avg=`grep 'avg:' ${sysbenchLog}/* | awk 'BEGIN{a=0}{a+=$2}END{print a/NR}'`

                echo ""
		
		totalCon=`expr $((${thrds}*${num}))`
                echo "Single tidb node sysbench test result as follows,Sysbench detailed result see dir ${sysbenchLog}"
                echo ""
                echo "Total concurrency ${totalCon} results:        "
                echo "----------------------------------------------"
		cat  ${sysbenchLog}/*.log | tail -26f

	fi
		

	if [ ! -d "${sysbenchDir}/results" ]; then
		mkdir -p ${sysbenchDir}/results
	fi

	echo ""
	echo "Start redirecting the above sort out results to ${sysbenchDir}/results/result.result"	
		
	echo "sysbenchLog dir name: ${sysbenchLog}" >> ${sysbenchDir}/results/result.result
	
        if [ ${nmonState} -eq 1 ]; then
                echo "nmonLog dir name: ${nmonDir}/nmonLog/${date}" >> ${sysbenchDir}/results/result.result
        fi
	
	echo "threads: ${totalCon}" >> ${sysbenchDir}/results/result.result
	echo "qps: ${qps}" >> ${sysbenchDir}/results/result.result
	echo "tps: ${tps}">> ${sysbenchDir}/results/result.result
	echo "latency-95: ${th}" >> ${sysbenchDir}/results/result.result
	echo "latency-min: ${min}">>${sysbenchDir}/results/result.result
	echo "latency-avg: ${avg}" >> ${sysbenchDir}/results/result.result
	echo "" >>${sysbenchDir}/results/result.result

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
	echo "Finished redirecting the above sort out results to ${sysbenchDir}/results/result.result"

	echo ""
		
	echo "Please check result output,Or manual view ${sysbenchDir}/results/result.result"
	echo ""
	cat ${sysbenchDir}/results/result.result	
}




###########################################################################
#	Sixth Part:	Get nmon log results                               |
#   function for get nmon log results                                      |
###########################################################################

function scpNmonLog(){
	echo ""
	echo "------------------------------------------------------------"
	echo "Nmon Log start scp from all tidb nodes to sysbench node dir."
	echo "------------------------------------------------------------"
   	
	if [ ! -d "${sysbenchDir}/nmonLog" ]; then
		mkdir -p ${sysbenchDir}/nmonLog/
	fi
	if [ ! -d "${sysbenchDir}/nmonLog/${date}" ]; then
                mkdir -p ${sysbenchDir}/nmonLog/${date}
        fi
	
	for ip in ${hostArry[*]}
   	do
		scp -r ${ip}:${nmonDir}/nmonLog/${date}/* ${sysbenchDir}/nmonLog/${date}/
	done


}



### Start Sysbench Test and Nmon 
function startSysbenchTest(){

	echo "************************ First *********************************"

	echo "   ---- Function genSysbenchConf start running ----"

	if [ ${nodeNums} -gt 1 ]; then
		# Query array portArry whether is existed repeat
		portArry1=($(echo ${portArry[*]} | sed 's/ /\n/g' | sort | uniq))
		portNums=${#portArry1[@]}
		if [ ${nodeNums} -eq ${portNums} ]; then
			genSysbenchConf
		else
						echo -e "\033[43;35m Found Error<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        echo -e "\033[43;35m Variable array portArry existed repeat,Please check repeat value.\033[0m \n"
                        echo -e "\033[43;35m Please Check<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        exit 1
		fi
	else
		genSysbenchConf
	fi
	
 	echo "   ---- Function genSysbenchConf has finished ----"	
	
	echo ""
        echo "************************ Second ********************************"
	
	echo "   ---- Function runTiDBTest start running ----"

	if [ ${nodeNums} -gt 1 ]; then
		# Query array portArry whether is existed repeat
		portArry1=($(echo ${portArry[*]} | sed 's/ /\n/g' | sort | uniq))
		portNums=${#portArry1[@]}
		if [ ${nodeNums} -eq ${portNums} ]; then
			runTiDBTest
		else
			echo -e "\033[43;35m Found Error<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        echo -e "\033[43;35m Variable array portArry existed repeat,Please check repeat value.\033[0m \n"
                        echo -e "\033[43;35m Please Check<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        exit 1
		fi
	else
		runTiDBTest
	fi
	
	echo "   ---- Function runTiDBTest has finished ----"
	
	echo ""

	echo "************************ Third *********************************"
	echo ""
	echo " ---- Start cycle check sysbench processes whether exited ----"
	
	# check sysbench processes if exited
	while true
	do
 		sysbenchPID=`ps -ef | grep sysbench | grep -v grep |grep -v sysbench.sh | awk '{print $2}'` 
        	if [ "$sysbenchPID" = "" ]; then
                	sleep 1
			echo ""
			echo "Cycle check sysbench processes not exited,Sysbench Test finished."
			break;
        	else
               		echo "runBenchmark.sh process exsits" > /dev/null 2>&1
        	fi
	done
	
	echo "   ---- Function getSysbenchTestResults start running -----"

	if [ ${nodeNums} -gt 1 ]; then
		# Query array portArry whether is existed repeat
		portArry1=($(echo ${portArry[*]} | sed 's/ /\n/g' | sort | uniq))
		portNums=${#portArry1[@]}
		if [ ${nodeNums} -eq ${portNums} ]; then
			getSysbenchTestResults
		else
			echo -e "\033[43;35m Found Error<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        echo -e "\033[43;35m Variable array portArry existed repeat,Please check repeat value.\033[0m \n"
                        echo -e "\033[43;35m Please Check<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\033[0m \n"
                        exit 1
		fi
	else
		getSysbenchTestResults
	fi
	echo "   ---- Function getSysbenchTestResults has finished ----"


	if [ ${nmonState} -eq 1 ]; then
		
        	echo ""
        	echo "************************ Fourth *********************************"
        	echo ""

        	echo "   ---- Function scpNmonLog start running ----"

		scpNmonLog
		
		echo ""
		echo "   ---- Function scpNmonLog has finished ----"
	else
		echo ""
		echo -e "\033[44;36m Nmon has disabeld,So don't need scp nmon log to sysbech dir. \033[0m \n"
	fi

	echo ""
	echo -e "\033[43;35m Congratulate Sysbench Finished. \033[0m \n"
}

### Stop Sysbench Test and Nmon 
function stopSysbenchTest(){
	
	# batch kill sysbench processes in all tidb nodes
	echo "#------------------ Start kill Sysbech Processes -----------------"
	echo ""
	result1=`ps -ef | grep sysbench | grep -v grep | grep -v sysbench.sh`
	if [ "${result1}" = "" ]; then
		echo "All sysbench test processes not running,Need not kill."
	else
		ps -ef | grep sysbench | grep -v grep | grep -v sysbench.sh |awk '{print $2}' | xargs sudo kill -9
                result=`ps -ef | grep sysbench | grep -v grep | grep -v sysbench.sh`
                if [ "${result}" = "" ]; then
                        echo "TiDB node ${ip} processes Sysbench has killed Success."
                else
                        echo -e "\033[43;35m TiDB node ${ip} processes Sysbench has killed Failed,Please check and manual kill. \033[0m \n"
                fi
	fi
		
	# batch kill nmon processes in all tidb nodes
	if [ ${nmonState} -eq 1 ]; then
		echo "#------------------ Start kill Nmon processes --------------------"
		echo ""
		for ip in ${hostArry[*]}
		do
			ssh ${ip} "ps -ef | grep nmon | grep -v grep | awk '{print $2}' | xargs sudo kill -9"
			result=`ssh ${ip} "ps -ef | grep nmon | grep -v grep"`
			if [ ! "$result" = "" ]; then
				echo -e "\033[43;35m TiDB node ${ip} processes Nmon has killed Failed,Please check and manual kill. \033[0m \n"
			else
				echo "TiDB node ${ip} processes Nmon has killed Success."
		
			fi
		done
	fi

        # kill sysbench.sh processes in the current node
	echo "#------------------ Start kill Sysbench.sh processes --------------------"
	echo ""
	ps -ef | grep sysbench.sh | grep -v grep | awk '{print $2}' | xargs sudo kill -9
	ps -ef | grep sysbench.sh | grep -v grep
	if [ $? -eq 0 ]; then
		echo -e "\033[43;35m Localhost processes sysbench.sh has killed Failed,Please check and manual kill. \033[0m \n"
	else
		echo "Localhost processes sysbench.sh has killed Success."
	fi

	
}


### Query sysbech and Nmon status
function statusSysbenchTest(){
	# query nodes sysbench processes
        echo "#--------------------------------------------------------------------------------"
        echo "                  Localhost Node Sysbench Processes                              "
        echo "#--------------------------------------------------------------------------------"
	echo ""

	ps -ef | grep sysbench | grep -v grep | grep -v sysbench.sh
	if [ $? -eq 0 ]; then
		echo "Localhost processes sysbench Existed and Running."
	else
		echo -e "\033[43;35m Localhost processes sysbench Not existed,Please check and manual query. \033[0m \n"
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
				echo -e "\033[43;38m Because variable nmonState equal 0,Show nmon need not turn on,That State is normal state. \033[0m \n" 
			else
				echo -e "\033[43;35m TiDB node ${ip} processes Nmon Not existed,Please manually check why nmon not be turned on,Or Manually start all tidb node nmon processed. \033[0m \n"
			fi
		fi
	done

	echo "#--------------------------------------------------------------------------------"
        echo "                      Localhost Sysbench.sh Processes                            "
        echo "#--------------------------------------------------------------------------------"
        echo ""

	ps -ef | grep sysbench.sh | grep -v grep
        if [ $? -eq 0 ]; then
                echo "Localhost script processes sysbench.sh Existed and Running."
        else
                echo -e "\033[43;35m Localhost script processes sysbench.sh Not existed,Please check and manual query. \033[0m \n"
        fi
}

### Main program entry
if [ ! -n "$1" ]; then
	echo ""
	echo -e "\033[43;35m Parameter can not be Null,Please input parameter start Or stop Or status. \033[0m \n"
	echo ""
else
	case "$1" in
		(start)
			startSysbenchTest
		;;
		(stop)
			stopSysbenchTest
		;;
		(status)
			statusSysbenchTest
		;;
	esac
fi
