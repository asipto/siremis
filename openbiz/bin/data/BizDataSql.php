<?PHP
/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin.data
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

/**
 * Class << BizDataSql >> is the class to constrcut SQL statement for BizDataObj
 *
 * @package openbiz.bin.data
 * @author Rocky Swen
 * @copyright Copyright (c) 2005-2009
 */
class BizDataSql
{
    protected $_tableColumns = null;
    protected $_tableJoins = null;
    protected $_joinAliasList = array();
    protected $_tableAliasList = array();
    protected $_sqlWhere = null;
    protected $_orderBy = null;
    protected $_otherSQL = null;
    protected $_aliasIndex = 0;
    protected $_mainTable;

    public function __construct()
    {
    }

    /**
     * Add main table in the sql statement T0 alias
     *
     * @param string $mainTable main table name
     * @return void
     **/
    public function addMainTable($mainTable)
    {
        $this->_mainTable = "$mainTable";
		$this->_tableJoins = " `$mainTable` T0 ";
    }

    /**
     * Add a join table in the sql statement Ti alias
     * <pre>
     *   SELECT T1.col, T2.col
     *   FROM table1 T1
     *       INNER JOIN table2 T2 ON T1.col1=T2.col1
     *       LEFT JOIN  table3 T3 ON T1.col1=T3.col1
     *   WHERE
     * </pre>
     *
     * @param TableJoin $tableJoin table join object
     * @return void
     **/
    public function addJoinTable($tableJoin)
    {
		$table = $tableJoin->getQuoted($tableJoin->m_Table);
        $joinType = $tableJoin->m_JoinType;
        $column = $tableJoin->m_Column;
        $joinRef = $tableJoin->m_JoinRef;
        $columnRef = $tableJoin->m_ColumnRef;

        $alias = "T".(count($this->_joinAliasList)+1);   // start with T1, T2
        $this->_joinAliasList[$tableJoin->m_Name] = $alias;
        $this->_tableAliasList[$table] = $alias;
        $aliasRef = $this->getJoinAlias($joinRef);
        $this->_tableJoins .= " $joinType $table $alias ON $alias.$column = $aliasRef.$columnRef ";
    }

    /**
     * Add a join table and cloumn in the sql statement
     *
     * @param string $join table join name
     * @param string $column column name
     * @return void
     **/
    public function addTableColumn($join, $column)
    {
        $tcol = $this->getTableColumn($join, $column);
        if (!$this->_tableColumns)
            $this->_tableColumns = $tcol;
        else
            $this->_tableColumns .= ", ".$tcol;
    }

    /**
     * Add SQL expression in the sql statement
     * sqlExpr has format of "...join1.column1, ... join2.column2...". Replace join with alias
     *
     * @param string $sqlExpr sql expression
     * @param string $alias sql alias
     * @return void
     **/
    public function addSqlExpression($sqlExpr, $alias=null)
    {
        if ($alias)
            $sqlExpr .= ' AS '.$alias;
        if (!$this->_tableColumns)
            $this->_tableColumns = $sqlExpr;
        else
            $this->_tableColumns .= ", ".$sqlExpr;
    }

    /**
     * Get join table alias
     *
     * @param string $join name of join
     * @return string join table alias
     */
    public function getJoinAlias($join)
    {
        if (!$join) // main table, no join
            return "T0";
        else
            return $this->_joinAliasList[$join];
    }

    /**
     * Get table column, combine a table with a column.
     *
     * @param string $join join name
     * @param string $col column
     * @return string table column combination string
     */
    public function getTableColumn($join, $col)
    {
        // check the function format on $col
        $alias = $this->getJoinAlias($join);
        return "$alias.$col";
    }

    /**
     * Reset SQL to be empty
     *
     * @return avoid
     */
    public function resetSQL()
    {
        $this->_sqlWhere = null;
        $this->_orderBy = null;
        $this->_otherSQL = null;
    }

    /**
     * Add the where clause (search rule) into the SQL statement
     *
     * @param string $sqlWhere SQL WHERE clause
     * @return void
     */
    public function addSqlWhere($sqlWhere)
    {
        if ($sqlWhere == null)
            return;
        if ($this->_sqlWhere == null)
        {
            $this->_sqlWhere = $sqlWhere;
        }
        elseif (strpos($this->_sqlWhere, $sqlWhere) === false)
        {
            $this->_sqlWhere .= " AND " . $sqlWhere;
        }
    }

    /**
     * Add order by clause
     *
     * @param string $orderBy SQL ORDER BY clause
     * @return void
     **/
    public function addOrderBy($orderBy)
    {
        if ($orderBy == null)
            return;
        if ($this->_orderBy == null)
        {
            $this->_orderBy = $orderBy;
        }
        elseif (strpos($this->_orderBy, $orderBy) === false)
        {
            $this->_orderBy .= " AND " . $orderBy;
        }
    }

    /**
     * Add other SQL clause
     *
     * @param string $otherSQL additional SQL statment
     * @return void
     **/
    public function addOtherSQL($otherSQL)
    {
        if ($otherSQL == null)
            return;
        if ($this->_otherSQL == null)
        {
            $this->_otherSQL = $otherSQL;
        }
        elseif (strpos($this->_otherSQL, $otherSQL) === false)
        {
            $this->_otherSQL .= " AND " . $otherSQL;
        }
    }

    /**
     * Add association in the SQL
     *
     * @param array $assoc additional SQL statment
     * @return void
     **/
    public function addAssociation($assoc)
    {
        $where = "";
        if ($assoc["Relationship"] == "1-M" || $assoc["Relationship"] == "M-1" || $assoc["Relationship"] == "1-1")
        {
            // assc table should same as maintable
            if ($assoc["Table"] != $this->_mainTable) return;
            // add table to join table
            $mytable_col = $this->getTableColumn(null, $assoc["Column"]);
            // construct table.column = 'field value'
            $where = $mytable_col." = '".$assoc["FieldRefVal"]."'";
        }
        elseif ($assoc["Relationship"] == "M-M")
        {
            // ... INNER JOIN xtable TX ON TX.column2 = T0.column
            // WHERE ... Tx.column1 = 'PrtTableColumnValue'
            $mytable_col = $this->getTableColumn(null, $assoc["Column"]);   // this table's alias.column
            $xtable = '`' . $assoc["XTable"] . '`';    // xtable's name
            $column1 = $assoc["XColumn1"]; // xtable column1
            $column2 = $assoc["XColumn2"]; // xtable column2
            $xalias = "TX";
            if (isset($this->_tableAliasList[$xtable]))
                $xalias = $this->_tableAliasList[$xtable];

            // add a table join for the xtable if such join of the table is not there before (note: may report error if has same join).
            //if (strpos($this->m_TableJoins, "JOIN $xtable") === false)
            if (!isset($this->_tableAliasList[$xtable]))
            {
                $this->_tableJoins .= " INNER JOIN $xtable $xalias ON $xalias.$column2 = $mytable_col ";
                $this->_tableAliasList[$xtable] = $xalias;
            }
            // add a new where condition
            $where = "$xalias.$column1 = '".$assoc["FieldRefVal"]."'";
        }

        if (strlen($where) > 1)
            $this->addSqlWhere($where);
    }

    /**
     * Get the SQL statement
     *
     * @return string SQL statement
     */
    public function getSqlStatement()
    {
        $ret = "SELECT " . $this->_tableColumns;
        $ret .= " FROM " . $this->_tableJoins;

        if ($this->_sqlWhere != null)
        {
            $ret .= " WHERE " . $this->_sqlWhere;
        }
        /*
        if ($this->m_OrderBy != null)
        {
            $ret .= " ORDER BY " . $this->m_OrderBy;
	}
        */
        if ($this->_otherSQL != null)
        {
            $ret .= " " . $this->_otherSQL;
        }
        if ($this->_orderBy != null)
        {
            $ret .= " ORDER BY " . $this->_orderBy;
        }
        return $ret;
    }
}
?>
