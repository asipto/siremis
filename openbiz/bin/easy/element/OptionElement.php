<?PHP
/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin.easy.element
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

include_once("InputElement.php");

/**
 * OptionElement is the base class of element that render list (from Selection.xml)
 * Used by :
 *   - {@link AutoSuggest}
 *   - {@link Checkbox}
 *   - {@link ColumnList}
 *   - {@link EditCombobox}
 *   - {@link LabelList}
 *   - {@link Listbox}
 *
 * @package openbiz.bin.easy.element
 * @author Rocky Swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class OptionElement extends InputElement
{
    public $m_SelectFrom;
    public $m_SelectFromSQL;
    public $m_SelectedList;

    /**
     * Read metadata info from metadata array and store to class variable
     *
     * @param array $xmlArr metadata array
     * @return void
     */
    protected function readMetaData(&$xmlArr)
    {
        parent::readMetaData($xmlArr);
        $this->m_SelectFrom = isset($xmlArr["ATTRIBUTES"]["SELECTFROM"]) ? $xmlArr["ATTRIBUTES"]["SELECTFROM"] : null;
        $this->m_SelectedList = isset($xmlArr["ATTRIBUTES"]["SELECTEDLIST"]) ? $xmlArr["ATTRIBUTES"]["SELECTEDLIST"] : null;
        $this->m_SelectFromSQL = isset($xmlArr["ATTRIBUTES"]["SELECTFROMSQL"]) ? $xmlArr["ATTRIBUTES"]["SELECTFROMSQL"] : null;
    }

    /**
     * Get select from
     *
     * @return string
     */
    protected function getSelectFrom()
    {
        $formobj = $this->getFormObj();
        return Expression::evaluateExpression($this->m_SelectFrom, $formobj);
    }

    protected function getSelectedList()
    {
        $formobj = $this->getFormObj();
        return Expression::evaluateExpression($this->m_SelectedList, $formobj);
    }
    
	protected function getSelectFromSQL()
    {
        $formobj = $this->getFormObj();
        return Expression::evaluateExpression($this->m_SelectFromSQL, $formobj);
    }

    /**
     * Render, draw the control according to the mode
     *
     * @return string HTML text
     */
    public function render()
    {
        return "";
    }

    /**
     * Get from list
     *
     * @param array $list
     * @return void
     */
    public function getFromList(&$list, $selectFrom=null)
    {
    	if (!$selectFrom) {
            $selectFrom = $this->getSelectFrom();
        }
        if (!$selectFrom) {
        	return $this->getSQLFromList($list);
        }
        $pos0 = strpos($selectFrom, "(");
        $pos1 = strpos($selectFrom, ")");
        if ($pos0>0 && $pos1 > $pos0)
        {  // select from xml file
            $xmlFile = substr($selectFrom, 0, $pos0);
            $tag = substr($selectFrom, $pos0 + 1, $pos1 - $pos0-1);
            $tag = strtoupper($tag);
            $xmlFile = BizSystem::GetXmlFileWithPath ($xmlFile);
            if (!$xmlFile) return;

            $xmlArr = &BizSystem::getXmlArray($xmlFile);
            if ($xmlArr)
            {
                $i = 0;
                if (!key_exists($tag, $xmlArr["SELECTION"]))
                    return;
                foreach($xmlArr["SELECTION"][$tag] as $node)
                {
                    $list[$i]['val'] = $node["ATTRIBUTES"]["VALUE"];
					if(isset($node["ATTRIBUTES"]["PICTURE"])) {
						$list[$i]['pic'] = $node["ATTRIBUTES"]["PICTURE"];
					}
                    if ($node["ATTRIBUTES"]["TEXT"])
                    {
                        $list[$i]['txt'] = $node["ATTRIBUTES"]["TEXT"];                        
                    }
                    else
                    {
                        $list[$i]['txt'] = $list[$i]['val'];
                    }
                    $i++;
                    
                }
                $this->translateList($list, $tag);	// supprot multi-language
            }
            return;
        }

        $pos0 = strpos($selectFrom, "[");
        $pos1 = strpos($selectFrom, "]");

        if ($pos0 > 0 && $pos1 > $pos0)
        {  // select from bizObj
            // support BizObjName[BizFieldName] or 
            // BizObjName[BizFieldName4Text:BizFieldName4Value] or 
            // BizObjName[BizFieldName4Text:BizFieldName4Value:BizFieldName4Pic]
            $bizObjName = substr($selectFrom, 0, $pos0);
            $pos3 = strpos($selectFrom, ":");
            if($pos3 > $pos0 && $pos3 < $pos1)
            {
                $fieldName = substr($selectFrom, $pos0 + 1, $pos3 - $pos0 - 1);
                $fieldName_v = substr($selectFrom, $pos3 + 1, $pos1 - $pos3 - 1);
            }
            else
            {
                $fieldName = substr($selectFrom, $pos0 + 1, $pos1 - $pos0 - 1);
                $fieldName_v = $fieldName;
            }
            $pos4 = strpos($fieldName_v, ":");
            if($pos4){
            	$fieldName_v_mixed = $fieldName_v;
            	$fieldName_v = substr($fieldName_v_mixed,0,$pos4);
            	$fieldName_p = substr($fieldName_v_mixed, $pos4+1, strlen($fieldName_v_mixed)-$pos4-1);
            	unset($fieldName_v_mixed);
            }
            $commaPos = strpos($selectFrom, ",", $pos1);
            if ($commaPos > $pos1)
                $searchRule = trim(substr($selectFrom, $commaPos + 1));
            $bizObj = BizSystem::getObject($bizObjName);
            if (!$bizObj)
                return;

            $recList = array();
            $oldAssoc = $bizObj->m_Association;
            $bizObj->m_Association = null;
            $recList = $bizObj->directFetch($searchRule);
            $bizObj->m_Association = $oldAssoc;

            foreach ($recList as $rec)
            {
                $list[$i]['val'] = $rec[$fieldName_v];
                $list[$i]['txt'] = $rec[$fieldName];
                $list[$i]['pic'] = $rec[$fieldName_p];
                $i++;
            }
            return;
        }

        // in case of a|b|c
        $recList = explode('|',$selectFrom);
        foreach ($recList as $rec)
        {
            $list[$i]['val'] = $rec;
            $list[$i]['txt'] = $rec;
            $list[$i]['pic'] = $rec;
            $i++;
        }
        return;
    }
    
    public function getSQLFromList(&$list)
    {
    	$sql = $this->getSelectFromSQL();
    	if (!$sql) return;
    	$formObj = $this->getFormObj();
    	$do = $formObj->getDataObj();
    	$db = $do->getDBConnection();
    	try {
    		$resultSet = $db->query($sql);
    		$recList = $resultSet->fetchAll();
	    	foreach ($recList as $rec)
	        {
	            $list[$i]['val'] = $rec[0];
	            $list[$i]['txt'] = isset($rec[1]) ? $rec[1] : $rec[0];
	            $i++;
	        }
    	}
    	catch (Exception $e)
        {
            BizSystem::log(LOG_ERR, "DATAOBJ", "Query Error: ".$e->getMessage());
            $this->m_ErrorMessage = "Error in SQL query: ".$sql.". ".$e->getMessage();
            throw new BDOException($this->m_ErrorMessage);
            return null;
        }
    }
    
    protected function translateList(&$list, $tag)
    {
    	$module = $this->getModuleName($this->m_FormName);
    	for ($i=0; $i<count($list); $i++)
    	{
    		$key = 'SELECTION_'.strtoupper($tag).'_'.$i.'_TEXT';
    		$list[$i]['txt'] = I18n::t($list[$i]['txt'], $key, $module);
    	}
    }
}

?>
