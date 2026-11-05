# Check period vesting account
echo "Check period vesting account"
PERIOD_VESTING_BALANCE_1=$(../build/terrad query bank spendable-balances terra1gyf58rxglrzp343d4wkw7vzlcw6d8knp2qmg0t --home ./chain-upgrade-data --output=json | jq '.balances[0].amount')
echo "PERIOD_VESTING_BALANCE_1 $PERIOD_VESTING_BALANCE_1"
sleep 5
PERIOD_VESTING_BALANCE_2="0"
echo "PERIOD_VESTING_BALANCE_2 $PERIOD_VESTING_BALANCE_2"
if (( ${PERIOD_VESTING_BALANCE_2//\"} <= ${PERIOD_VESTING_BALANCE_1//\"} )); then
    echo "Period vesting account balance must be updated every block"
    exit 1
fi