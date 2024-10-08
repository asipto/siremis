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

/**
 * Element class is the base class of all HTML Element
 *
 * @package openbiz.bin.easy.element
 * @author Rocky Swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class Element extends MetaObject implements iUIControl
{
    public $m_Style;
    public $m_cssClass;
    public $m_cssErrorClass;
    public $m_Width;
    public $m_Height;
    public $m_Hidden = "N";       // support expression
    public $m_HTMLAttr;
    public $m_Label;
    public $m_EventHandlers;
    public $m_Translatable;
    public $m_FuzzySearch;
    public $m_OnEventLog;
    public $m_AllowURLParam = 'N';

    public $m_SortFlag;
    public $m_Value = "";
    public $m_FormName;
    public $m_ElementSet;
    public $m_Required = null;
    public $m_Validator = null;
    public $m_ClientValidator = null;
	public $m_KeepCookie = null;
	public $m_CookieLifetime = 3600;
	
    /**
     * Initialize Element with xml array
     *
     * @param array $xmlArr
     * @return void
     */
    function __construct(&$xmlArr, $formObj)
    {
        $this->m_FormName = $formObj->m_Name;
        $this->m_Package = $formObj->m_Package;

        $this->readMetaData($xmlArr);
                
        $this->translate();	// translate for multi-language support
    }

    /**
     * Read array meta data, and store to meta object
     *
     * @param array $xmlArr
     * @return void
     */
    protected function readMetaData(&$xmlArr)
    {
        $this->m_Name = isset($xmlArr["ATTRIBUTES"]["NAME"]) ? $xmlArr["ATTRIBUTES"]["NAME"] : null;
        $this->m_Class = isset($xmlArr["ATTRIBUTES"]["CLASS"]) ? $xmlArr["ATTRIBUTES"]["CLASS"] : null;
        $this->m_Description = isset($xmlArr["ATTRIBUTES"]["DESCRIPTION"]) ? $xmlArr["ATTRIBUTES"]["DESCRIPTION"] : null;
        $this->m_Access = isset($xmlArr["ATTRIBUTES"]["ACCESS"]) ? $xmlArr["ATTRIBUTES"]["ACCESS"] : null;
        $this->m_DefaultValue = isset($xmlArr["ATTRIBUTES"]["DEFAULTVALUE"]) ? $xmlArr["ATTRIBUTES"]["DEFAULTVALUE"] : null;
        $this->m_cssClass = isset($xmlArr["ATTRIBUTES"]["CSSCLASS"]) ? $xmlArr["ATTRIBUTES"]["CSSCLASS"] : null;
        $this->m_cssErrorClass = isset($xmlArr["ATTRIBUTES"]["CSSERRORCLASS"]) ? $xmlArr["ATTRIBUTES"]["CSSERRORCLASS"] : "input_error";
        $this->m_Style = isset($xmlArr["ATTRIBUTES"]["STYLE"]) ? $xmlArr["ATTRIBUTES"]["STYLE"] : null;
        $this->m_Width = isset($xmlArr["ATTRIBUTES"]["WIDTH"]) ? $xmlArr["ATTRIBUTES"]["WIDTH"] : null;
        $this->m_Height = isset($xmlArr["ATTRIBUTES"]["HEIGHT"]) ? $xmlArr["ATTRIBUTES"]["HEIGHT"] : null;
        $this->m_Hidden = isset($xmlArr["ATTRIBUTES"]["HIDDEN"]) ? $xmlArr["ATTRIBUTES"]["HIDDEN"] : null;
        $this->m_HTMLAttr = isset($xmlArr["ATTRIBUTES"]["HTMLATTR"]) ? $xmlArr["ATTRIBUTES"]["HTMLATTR"] : null;
        $this->m_ElementSet = isset($xmlArr["ATTRIBUTES"]["ELEMENTSET"]) ? $xmlArr["ATTRIBUTES"]["ELEMENTSET"] : null;        
        $this->m_Text = isset($xmlArr["ATTRIBUTES"]["TEXT"]) ? $xmlArr["ATTRIBUTES"]["TEXT"] : null;
        $this->m_Translatable = isset($xmlArr["ATTRIBUTES"]["TRANSLATABLE"]) ? $xmlArr["ATTRIBUTES"]["TRANSLATABLE"] : null;
        $this->m_FuzzySearch = isset($xmlArr["ATTRIBUTES"]["FUZZYSEARCH"]) ? $xmlArr["ATTRIBUTES"]["FUZZYSEARCH"] : null;
        $this->m_OnEventLog = isset($xmlArr["ATTRIBUTES"]["ONEVENTLOG"]) ? $xmlArr["ATTRIBUTES"]["ONEVENTLOG"] : null;
        $this->m_Required = isset($xmlArr["ATTRIBUTES"]["REQUIRED"]) ? $xmlArr["ATTRIBUTES"]["REQUIRED"] : null;
        $this->m_Validator = isset($xmlArr["ATTRIBUTES"]["VALIDATOR"]) ? $xmlArr["ATTRIBUTES"]["VALIDATOR"] : null;
        $this->m_ClientValidator = isset($xmlArr["ATTRIBUTES"]["CLIENTVALIDATOR"]) ? $xmlArr["ATTRIBUTES"]["CLIENTVALIDATOR"] : null;
        $this->m_AllowURLParam = isset($xmlArr["ATTRIBUTES"]["ALLOWURLPARAM"]) ? $xmlArr["ATTRIBUTES"]["ALLOWURLPARAM"] : 'Y';
        $this->m_KeepCookie = isset($xmlArr["ATTRIBUTES"]["KEEPCOOKIE"]) ? $xmlArr["ATTRIBUTES"]["KEEPCOOKIE"] : 'N';
        $this->m_CookieLifetime = isset($xmlArr["ATTRIBUTES"]["COOKIELIFETIME"]) ? (int)$xmlArr["ATTRIBUTES"]["COOKIELIFETIME"] : '3600';

        // read EventHandler element
        if (isset($xmlArr["EVENTHANDLER"]))  // 2.1 eventhanlders
            $this->m_EventHandlers = new MetaIterator($xmlArr["EVENTHANDLER"],"EventHandler");

        if ($this->m_EventHandlers != null)
        {
            foreach ($this->m_EventHandlers as $eventHandler)
                $eventHandler->setFormName($this->m_FormName, $this->m_Name);
        }

        // add description and client validator on HTMLAttr
        $this->m_HTMLAttr .= " title='".$this->m_Description."'"." clientValidator='".$this->m_ClientValidator."'";
    }

    /**
     * Get form ({@link EasyForm}) object
     *
     * @return EasyForm
     */
    protected function getFormObj()
    {
        return BizSystem::objectFactory()->getObject($this->m_FormName);
    }

    //
    /**
     * Adjust form ({@link EasyForm}) name
     * change the form name after inherit from parent form
     *
     * @param string $formName
     * @return void
     */
    public function adjustFormName($formName)
    {
        if ($this->m_FormName == $formName)
            return;
        $this->m_FormName = $formName;
        if ($this->m_EventHandlers != null)
        {
            foreach ($this->m_EventHandlers as $eventHandler)
                $eventHandler->adjustFormName($this->m_FormName);
        }
    }

    public function reset()
    {
    	$this->m_Value = null;
    	if ($this->m_EventHandlers != null)
        {
            foreach ($this->m_EventHandlers as $eventHandler)
                $eventHandler->m_FormedFunction = null;
        }
    }

    /**
     * Set value of element
     *
     * @param mixed $value
     */
    public function setValue($value)
    {
        $this->m_Value = $value;        
        if($this->m_KeepCookie=='Y'){
        	if($value!=""){
        		$formName = $this->getFormObj()->m_Name;       
        		setcookie($formName."-".$this->m_Name,$value,time()+(int)$this->m_CookieLifetime,"/");
        	}
        }
    }

    /**
     * Get value of element
     *
     * @return mixed
     */
    public function getValue()
    {
        return $this->m_Value;
    }

    /**
     * Get property of element
     *
     * @param string $propertyName
     * @return mixed
     */
    public function getProperty($propertyName)
    {
        if ($propertyName == "Value") return $this->m_Value;
        $ret = parent::getProperty($propertyName);
        if ($ret) return $ret;
        return $this->$propertyName;
    }

    /**
     * Check if the element can be displayed
     *
     * @return boolean
     */
    public function canDisplayed()
    {
        if (!$this->allowAccess())
            return false;
        if ($this->getHidden() == "Y")
            return false;
        return true;
    }

    /**
     * Get default value
     *
     * @return mixed
     */
    public function getDefaultValue()
    {
        if ($this->m_DefaultValue == "" && $this->m_KeepCookie!='Y')
            return "";
        $formObj = $this->getFormObj();
        if($this->m_KeepCookie=='Y'){
        	$cookieName = $formObj->m_Name."-".$this->m_Name;      
        	$cookieName = str_replace(".","_",$cookieName);
        	$defValue = $_COOKIE[$cookieName];        	
        }                
        if(!$defValue){
        	$defValue = Expression::evaluateExpression($this->m_DefaultValue, $formObj);
        }
        return $defValue;
    }

    /**
     * Render the element by html
     *
     * @return string HTML text
     */
    public function render()
    {
     	return "";
    }

    /**
     * Get hidden status
     *
     * @return mixed
     */
    protected function getHidden()
    {
        $formObj = $this->getFormObj();
        return Expression::evaluateExpression($this->m_Hidden, $formObj);
    }

    /**
     * Get style of element
     *
     * @return string
     */
    protected function getStyle()
    {        
		$formobj = $this->getFormObj();    	
        $htmlClass = Expression::evaluateExpression($this->m_cssClass, $formobj);
        $htmlClass = "CLASS='$htmlClass'";
        if(!$htmlClass){
        	$htmlClass = null;
        }
        $style ='';
        if ($this->m_Width && $this->m_Width>=0)
            $style .= "width:".$this->m_Width."px;";
        if ($this->m_Height && $this->m_Height>=0)
            $style .= "height:".$this->m_Height."px;";
        if ($this->m_Style)
            $style .= $this->m_Style;
        if (!isset($style) && !$htmlClass)
            return null;
        if (isset($style))
        {
            
            $style = Expression::evaluateExpression($style, $formobj);
            $style = "STYLE='$style'";
        }
        if(isset($formobj->m_Errors[$this->m_Name]))
        {
      	    $htmlClass = "CLASS='".$this->m_cssErrorClass."'";
        }
        if ($htmlClass)
            $style = $htmlClass." ".$style;
        return $style;
    }

    /**
     * Get text of element
     *
     * @return string
     */
    protected function getText()
    {
        if ($this->m_Text == null)
            return null;
        $formobj = $this->getFormObj();
        return Expression::evaluateExpression($this->m_Text, $formobj);
    }

    /**
     * Get shortcut key function map
     *
     * @return array
     */
    public function getSCKeyFuncMap()
    {
        if (!$this->canDisplayed()) return null;

        $map = array();
        /**
         * @todo need to remove, not used (mr_a_ton)
         */
        //$formObj = $this->getFormObj(); // not used

        if ($this->m_EventHandlers == null)
            return null;
        foreach ($this->m_EventHandlers as $eventHandler)
        {
            if ($eventHandler->m_ShortcutKey)
            {
                $map[$eventHandler->m_ShortcutKey] = $eventHandler->getFormedFunction();
            }
        }
        return $map;
    }

    /**
     * Get context menu
     *
     * @return array
     */
    public function getContextMenu()
    {
        if (!$this->canDisplayed()) return null;
        $menus = array();
        $formObj = $this->getFormObj();
        if ($this->m_EventHandlers == null)
            return null;
        $i = 0;
        foreach ($this->m_EventHandlers as $eventHandler)
        {
            if ($eventHandler->m_ContextMenu)
            {
                $menus[$i]['text'] = $eventHandler->m_ContextMenu;
                $menus[$i]['func'] = $eventHandler->getFormedFunction();
                $menus[$i]['key']  = $eventHandler->m_ShortcutKey;
            }
            $i++;
        }
        return $menus;
    }

    /**
     * Get function of element in JavaScript format
     *
     * @return string
     */
    protected function getFunction()
    {
        $name = $this->m_Name;
        // loop through the event handlers
        $func = "";

        if ($this->m_EventHandlers == null)
            return null;
        $formobj = $this->getFormObj();
        foreach ($this->m_EventHandlers as $eventHandler)
        {
            $ehName = $eventHandler->m_Name;
            $event = $eventHandler->m_Event;
            $type = $eventHandler->m_FunctionType;
            if (!$event) continue;
            if(isset($events) && isset($events[$event])){
            	$events[$event]=array_merge(array($events[$event]),array($eventHandler->getFormedFunction()));
            }else{
            	$events[$event]=$eventHandler->getFormedFunction();
            }

        }
		foreach ($events as $event=>$function){
			if(is_array($function)){
				foreach($function as $f){
					$function_str.=$f.";";
				}
				$func .= " $event=\"$function_str\"";
			}else{
				$func .= " $event=\"$function\"";
			}
		}
        return $func;
    }
    
    public function getFunctionByEventHandlerName($eventHandlerName)
    {
    	if ($this->m_EventHandlers == null)
            return null;
    	$eventHandler = $this->m_EventHandlers->get($eventHandlerName);
    	if ($eventHandler)
    		$func = Expression::evaluateExpression($eventHandler->m_Function, $formobj);
    	return $func;
    }

    /**
     * Get redirect page
     *
     * @param string $eventHandlerName event handler name
     * @return string
     */
    public function getRedirectPage($eventHandlerName)
    {
        $formObj = $this->getFormObj();
        $eventHandler = $this->m_EventHandlers->get($eventHandlerName);
        if (!$eventHandler) return null;
        //echo $evthandler->m_RedirectPage."<br>";
        return Expression::evaluateExpression($eventHandler->m_RedirectPage, $formObj);
    }

    /**
     * Get parameter
     *
     * @param string $eventHandlerName
     * @return mixed
     */
    public function getParameter($eventHandlerName){
    	$formObj = $this->getFormObj();
        $eventHandler = $this->m_EventHandlers->get($eventHandlerName);
        if (!$eventHandler) return null;
        //echo $evthandler->m_RedirectPage."<br>";
        return Expression::evaluateExpression($eventHandler->m_Parameter, $formObj);
    }
    
    /**
     * Get function type
     *
     * @param string $eventHandlerName event handler name
     * @return string function type in string format
     */
    public function getFunctionType($eventHandlerName)
    {
        $eventHandler = $this->m_EventHandlers->get($eventHandlerName);
        if (!$eventHandler) return null;
        return $eventHandler->m_FunctionType;
    }

    /**
     * Check if element must required (must have value)
     *
     * @return boolean
     */
    public function checkRequired()
    {
        if (!$this->m_Required || $this->m_Required == "")
            return false;
        else if ($this->m_Required == "Y")
            $required = true;
        else if($required != "N")
            $required = false;
        else
            $required = Expression::evaluateExpression($this->m_Required, $this->getFormObj());

        return $required;
    }

    /**
     * Validate element
     * 
     * @return boolean
     */
    public function validate()
    {
        $ret = true;
        if ($this->m_Validator)
            $ret = Expression::evaluateExpression($this->m_Validator, $this->getFormObj());
        return $ret;
    }

    /**
     * Get client validator
     *
     * @return string
     */
    public function getClientValidator()
    {
        if ($this->m_ClientValidator)
            return $this->m_ClientValidator;

        //return Expression::evaluateExpression($this->m_ClientValidator, $this->getFormObj());
        return null;
    }
    
    public function matchRemoteMethod($method)
    {
        return false;
    }
    
    protected function translate()
    {
    	$module = $this->getModuleName($this->m_FormName);
    	if (!empty($this->m_Text))
    		$this->m_Text = I18n::t($this->m_Text, $this->getTransKey('Text'), $module);
    	if (!empty($this->m_Label))
    		$this->m_Label = I18n::t($this->m_Label, $this->getTransKey('Label'), $module);
    	if (!empty($this->m_Description))
    		$this->m_Description = I18n::t($this->m_Description, $this->getTransKey('Description'), $module);
    }
    
    protected function getTransKey($name)
    {
    	$shortFormName = substr($this->m_FormName,intval(strrpos($this->m_FormName,'.')+1));
    	return strtoupper($shortFormName.'_'.$this->m_Name.'_'.$name);
    }
}

/**
 * EventHandler class is manage event handler of element
 *
 * @package openbiz.bin.easy.element
 * @author Rocky Swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class EventHandler
{
    public $m_Name;
    public $m_Event;
    public $m_Function;     // support expression
    public $m_FunctionType;
    public $m_PostAction;   // support expression
    public $m_ShortcutKey;
    public $m_ContextMenu;
    public $m_RedirectPage;
    public $m_Parameter;
    public $m_EventLogMsg;
    public $m_FormedFunction;
    private $_formName;
    private $_elemName;

    // add URL here so that direct url string can be given
    public $m_URL;

    /**
     * Initialize EventHandler with xml array
     *
     * @param array $xmlArr
     * @return void
     */
    function __construct(&$xmlArr)
    {
        $this->m_Name = isset($xmlArr["ATTRIBUTES"]["NAME"]) ? $xmlArr["ATTRIBUTES"]["NAME"] : null;
        $this->m_Event = isset($xmlArr["ATTRIBUTES"]["EVENT"]) ? $xmlArr["ATTRIBUTES"]["EVENT"] : null;
        $this->m_Function = isset($xmlArr["ATTRIBUTES"]["FUNCTION"]) ? $xmlArr["ATTRIBUTES"]["FUNCTION"] : null;
        $this->m_OrigFunction = $this->m_Function;
        $this->m_FunctionType = isset($xmlArr["ATTRIBUTES"]["FUNCTIONTYPE"]) ? $xmlArr["ATTRIBUTES"]["FUNCTIONTYPE"] : null;
        $this->m_PostAction = isset($xmlArr["ATTRIBUTES"]["POSTACTION"]) ? $xmlArr["ATTRIBUTES"]["POSTACTION"] : null;
        $this->m_ShortcutKey = isset($xmlArr["ATTRIBUTES"]["SHORTCUTKEY"]) ? $xmlArr["ATTRIBUTES"]["SHORTCUTKEY"] : null;
        $this->m_ContextMenu = isset($xmlArr["ATTRIBUTES"]["CONTEXTMENU"]) ? $xmlArr["ATTRIBUTES"]["CONTEXTMENU"] : null;
        $this->m_RedirectPage = isset($xmlArr["ATTRIBUTES"]["REDIRECTPAGE"]) ? $xmlArr["ATTRIBUTES"]["REDIRECTPAGE"] : null;        
		$this->m_Parameter = isset($xmlArr["ATTRIBUTES"]["PARAMETER"]) ? $xmlArr["ATTRIBUTES"]["PARAMETER"] : null;        
        $this->m_EventLogMsg = isset($xmlArr["ATTRIBUTES"]["EVENTLOGMSG"]) ? $xmlArr["ATTRIBUTES"]["EVENTLOGMSG"] : null;
        $this->m_URL = isset($xmlArr["ATTRIBUTES"]["URL"]) ? $xmlArr["ATTRIBUTES"]["URL"] : null;
    }

    /**
     * Set form name that contain element and EventHandler
     * 
     * @param string $formName
     * @param string $elemName
     * @return void
     */
    public function setFormName($formName, $elemName)
    {
        $this->_formName = $formName;
        $this->_elemName = $elemName;
        if (strpos($this->m_Function, "js:")===0)
            return;
        // if no class name, add default class name. i.e. NewRecord => ObjName.NewRecord
        if ($this->m_Function)
        {
            $pos_dot = strpos($this->m_Function, ".");
            $pos_lpt = strpos($this->m_Function, "(");
            if (!$pos_dot || $pos_lpt < $pos_dot)
                $this->m_Function = $this->_formName.".".$this->m_Function;
        }
        $this->translate();	// translate for multi-language support
    }

    /**
     * Adjust form name
     *
     * @param string $formName
     * @return void
     */
    public function adjustFormName($formName)
    {
        $this->_formName = $formName;
        // if no class name, add default class name. i.e. NewRecord => ObjName.NewRecord
        if ($this->m_Function)
        {
            $pos = strrpos($this->m_Function, ".");
            if ($pos > 0)
                $this->m_Function = $this->_formName.".".substr($this->m_Function, $pos+1);
        }
    }

    /**
     * Get formed function
     *
     * @return string
     */
    public function getFormedFunction()
    {
        //return $this->getInvokeAction();
        $name = $this->_elemName;
        $ehName = $this->m_Name;
        $formobj = BizSystem::objectFactory()->getObject($this->_formName);
        if ($this->m_FormedFunction)
        {
            return $this->m_FormedFunction;
        }        
        if (!$this->m_FormedFunction || $isDataPanelElement==true)
        {
            // add direct URL support
            if ($this->m_URL) 
            {
                $_func = "loadPage('" . $this->m_URL . "');";
                $_func = Expression::evaluateExpression($_func, $formobj);
            }
            else if (strpos($this->m_Function, "js:") === 0) 
            {
                $_func = substr($this->m_Function, 3).";";
                $_func = Expression::evaluateExpression($_func, $formobj);
            }
            else 
            {
                //$temp = ($this->m_FunctionType==null) ? "" : ",'".$this->m_FunctionType."'";
                //$_func = "SetOnElement('$name:$ehName'); $selectRecord CallFunction('" . $this->m_Function . "'$temp);";
                //$_func = "Openbiz.CallFunction('" . $this->m_Function . "'$temp);";
                $_func = Expression::evaluateExpression($this->m_Function, $formobj);
                $options = "{'type':'$this->m_FunctionType','target':'','evthdl':'$name:$ehName'}";
                $_func = "Openbiz.CallFunction('$_func',$options);";
            }
            $this->m_FormedFunction = $_func;
        }
        return $this->m_FormedFunction;
    }
    
    public function getInvokeAction()
    {
        if ($this->m_FormedFunction)
            return $this->m_FormedFunction;
    	$name = $this->_elemName;
        $ehName = $this->m_Name;
        $formobj = BizSystem::objectFactory()->getObject($this->_formName);
     
        if (!$this->m_FormedFunction)
        {
            // add direct URL support
            if ($this->m_URL)
                $_func = "loadPage('" . $this->m_URL . "');";
            else if (strpos($this->m_Function, "js:") === 0)
                $_func = substr($this->m_Function, 3).";";
            else
            {
                $temp = ($this->m_FunctionType==null) ? "" : ",'".$this->m_FunctionType."'";                
                //$_func = "SetOnElement('$name:$ehName'); Openbiz.CallFunction('" . $this->m_Function . "'$temp);";
                list($funcName, $funcParams) = $this->parseFunction($this->m_Function);
                $funcParams = Expression::evaluateExpression($funcParams, $formobj);
                $action = "$name:$ehName";
                // TODO: encrypt paramString to add more security
                $_func = "Openbiz.invoke('$this->_formName','$action','$funcParams'$temp);";
            }
            //$_func = Expression::evaluateExpression($_func, $formobj);
            $this->m_FormedFunction = $_func;
        }
        return $this->m_FormedFunction;
    }
    
    // parse function string and get functionName and functionParams
    public function parseFunction($funcString)
    {
        $pos = strpos($funcString, "(");
        $pos1 = strpos($funcString, ")");
        if ($pos>0 && $pos1>$pos)
        {
            $funcName = substr($funcString,0,$pos);
            $funcParams = substr($funcString,$pos+1,$pos1-$pos-1);
            return array($funcName, $funcParams);
        }
        return null;
    }
    
    protected function translate()
    {
    	$module = substr($this->_formName,0,intval(strpos($this->_formName,'.')));
    	if (!empty($this->m_ContextMenu))
    		$this->m_ContextMenu = I18n::t($this->m_ContextMenu, $this->getTransKey('ContextMenu'), $module);
    }
    
    protected function getTransKey($name)
    {
    	$shortFormName = substr($this->m_FormName,intval(strrpos($this->m_FormName,'.'))+1);
    	return strtoupper($shortFormName.'_'.$this->m_Name.'_'.$name);
    }
}
?>
