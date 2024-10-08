<?PHP
/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin.easy
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

/**
 * EasyView class is the class that contains list of forms.
 * View is same as html page.
 *
 * @package openbiz.bin.easy
 * @author rocky swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class EasyView extends MetaObject implements iSessionObject
{
    public $m_Title;
    public $m_Keywords;
	public $m_TemplateEngine;
    public $m_TemplateFile;
    public $m_ViewSet;
    public $m_Tab;
    public $m_FormRefs;
	public $m_Tiles;

    public $m_IsPopup = false;
    public $m_Height;
    public $m_Width;
    public $m_ConsoleOutput = true;

    public $m_MessageFile = null;        // message file path
    protected $m_Messages;
    public $m_CacheLifeTime = 0;

    /**
     * Initialize EasyView with xml array
     *
     * @param array $xmlArr
     * @return void
     */
    public function __construct(&$xmlArr)
    {
        $this->readMetadata($xmlArr);
    }

    /**
     * Read Metadata from xml array
     * 
     * @param array $xmlArr
     * @return void
     */
    protected function readMetadata(&$xmlArr)
    {
        parent::readMetaData($xmlArr);
        $this->m_Name = $this->prefixPackage($this->m_Name);
        $this->m_Title = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["TITLE"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["TITLE"] : null;
        $this->m_Keywords = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["KEYWORDS"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["KEYWORDS"] : null;
        $this->m_TemplateEngine = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["TEMPLATEENGINE"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["TEMPLATEENGINE"] : null;
        $this->m_TemplateFile = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["TEMPLATEFILE"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["TEMPLATEFILE"] : null;
        $this->m_ViewSet = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["VIEWSET"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["VIEWSET"] : null;
        $this->m_Tab = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["TAB"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["TAB"] : null;

        $this->m_FormRefs = new MetaIterator($xmlArr["EASYVIEW"]["FORMREFERENCES"]["REFERENCE"],"FormReference",$this);
        if(isset($xmlArr["EASYVIEW"]["FORMREFERENCELIBS"]))
        {
        	$this->m_FormRefLibs = new MetaIterator($xmlArr["EASYVIEW"]["FORMREFERENCELIBS"]["REFERENCE"],"FormReference",$this);
        }
        $this->m_MessageFile = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["MESSAGEFILE"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["MESSAGEFILE"] : null;
        $this->m_Messages = Resource::loadMessage($this->m_MessageFile);
        $this->m_CacheLifeTime = isset($xmlArr["EASYVIEW"]["ATTRIBUTES"]["CACHELIFETIME"]) ? $xmlArr["EASYVIEW"]["ATTRIBUTES"]["CACHELIFETIME"] : "0";
		
        $this->readTile($xmlArr);	// TODO: is this needed as title supports expression?
        
        if (empty($this->m_Title))
        	$this->m_Title = $this->m_Description;
        $this->translate();	// translate for multi-language support
    }
    
    protected function readTile(&$xmlArr)
    {
    	if (isset($xmlArr["EASYVIEW"]["TILE"]))
        {
        	$this->m_FormRefs = array();
        	if (isset($xmlArr["EASYVIEW"]["TILE"]["ATTRIBUTES"])) 
        	{
        		$tileName = $xmlArr["EASYVIEW"]["TILE"]["ATTRIBUTES"]["NAME"];
        		$this->m_Tiles[$tileName] = new MetaIterator($xmlArr["EASYVIEW"]["TILE"]["REFERENCE"],"FormReference",$this);
        	}
        	else 
        	{
        		foreach ($xmlArr["EASYVIEW"]["TILE"] as $child)
        		{
        			$tileName = $child["ATTRIBUTES"]["NAME"];
	        		$this->m_Tiles[$tileName] = new MetaIterator($child["REFERENCE"],"FormReference",$this);
        		}
        	}
        	//echo "<pre>"; print_r($this->m_Tiles); echo "</pre>"; 
        	$tmp = array();
        	$this->m_FormRefs = new MetaIterator($tmp,"",$this);
        	foreach ($this->m_Tiles as $tile)
        	{
        		foreach ($tile as $ref)
        			$this->m_FormRefs->set($ref->m_Name, $ref);
        	}
        }
    }

    /**
     * Check the Form is in the lib
     *
     * @param string $formName form name     
     * @return bool inside or not
     */
    public function isInFormRefLibs($formName)
    {
    	if($this->m_FormRefLibs)
    	{
    		$this->m_FormRefLibs->rewind();
    		while($this->m_FormRefLibs->valid())
    		{
    			$reference = $this->m_FormRefLibs->current();
    			if($reference->m_Name == $formName)
    			{
    				return true;
    			}
    			$this->m_FormRefLibs->next();
    		}
    		return false;
    	}
    	else
    	{
    		return true;
    	}
    }
    
    /**
     * Get message, and translate it
     *
     * @param <type> $msgId message Id
     * @param array $params
     * @return string message string
     */
    protected function getMessage($msgId, $params=array())
    {
        $message = isset($this->m_Messages[$msgId]) ? $this->m_Messages[$msgId] : constant($msgId);
        //$message = I18n::getInstance()->translate($message);
        $message = I18n::t($message, $messageId, $this->getModuleName($this->m_Name));
        return vsprintf($message,$params);
    }


    /**
     * Get/Retrieve Session data of this object
     *
     * @param SessionContext $sessionContext
     * @return void
     */
    public function getSessionVars($sessionContext)
    {
    }
    
    /**
     * Save Session data of this object
     *
     * @param SessionContext $sessionContext
     * @return void
     */
    public function setSessionVars($sessionContext)
    {
    }

    /**
     * Get view set name
     *
     * @return mixed viewset name or null
     */
    public function getViewSet()
    { 
        return $this->m_ViewSet;
    }

    /**
     * Set the Render output to console (as calling print ...) or to a string buffer
     *
     * @param boolean $consoleOutput
     * @return void
     */
    public function setConsoleOutput($consoleOutput)
    {
        $this->m_ConsoleOutput = $consoleOutput;
    }

    /**
     * Proses rule
     *
     * @return void
     */
    public function processRule()
    {}

    /**
     * Set parameters
     *
     * @return void
     */
    public function setParameters()
    {}

    /**
     * Render this view.
     *
     * @return mixed either print html content, or return html content
     */
    public function render()
    {
    	if (!$this->allowAccess())
        {
            $accessDenyView = BizSystem::objectFactory()->getObject(ACCESS_DENIED_VIEW);
            return $accessDenyView->render();
        }

        $this->initAllForms();

        // check the "fld_..." arg in url and put it in the search rule
        $this->processRequest();

        return $this->_render();
    }

    /**
     * Render this view. This function is called by Render() or ReRender()
     *
     * @return mixed either print html content or return html content if called by Render(), or void if called by ReRender()
     */
    protected function _render()
    {
        $this->setClientScripts();

        if($this->m_CacheLifeTime>0)
        {
            $pageUrl = $this->curPageURL();
            $cache_id = md5($pageUrl);
            //try to process cache service.
            $cacheSvc = BizSystem::getService(CACHE_SERVICE,1);
            $cacheSvc->init($this->m_Name,$this->m_CacheLifeTime);
            if($cacheSvc->test($cache_id))
            {
                BizSystem::log(LOG_DEBUG, "VIEW", "Cache Hit. url = ".$pageUrl);
                $output = $cacheSvc->load($cache_id);
            }
            else
            {
                include_once(OPENBIZ_BIN."/easy/ViewRenderer.php");
                $this->m_ConsoleOutput = false;
                $output = ViewRenderer::render($this);
                BizSystem::log(LOG_DEBUG, "VIEW", "Set cache. url = ".$pageUrl);
                $cacheSvc->save($output, $cache_id);
            }
            print $output;
        }
        else
        {
            include_once(OPENBIZ_BIN."/easy/ViewRenderer.php");
            ViewRenderer::render($this);
        }
        return;
        /*
        $this->setClientScripts();
      	include_once(OPENBIZ_BIN."/easy/ViewRenderer.php"); 
	    return ViewRenderer::render($this);*/
    }

    /**
     * Get current page URL
     * NOTE:
     * This method on next version maybe removed.
     * New method is {@link getCurrentPageUrl}
     * 
     * @return string page URL
     */
    public function curPageURL()
    {
        return $this->getCurrentPageUrl();
    }

    /**
     * Get current page URL
     *
     * @return string page URL
     */

    public function getCurrentPageUrl()
    {
        $pageURL = 'http';
        if ($_SERVER["HTTPS"] == "on")
        {
            $pageURL .= "s";
        }
        $pageURL .= "://";
        if ($_SERVER["SERVER_PORT"] != "80")
        {
            $pageURL .= $_SERVER["SERVER_NAME"].":".$_SERVER["SERVER_PORT"].$_SERVER["REQUEST_URI"];
        }
        else
        {
            $pageURL .= $_SERVER["SERVER_NAME"].$_SERVER["REQUEST_URI"];
        }
        return $pageURL;
    }

    /**
     * Set default client javascript and css that included in the html content
     *
     * @return void
     */
    protected function setClientScripts()
    {
        BizSystem::clientProxy()->appendScripts("prototype", "prototype.js");
        BizSystem::clientProxy()->appendScripts("scriptaculous", "scriptaculous.js");
        BizSystem::clientProxy()->appendScripts("openbiz", "openbiz.js");      
        BizSystem::clientProxy()->appendStyles("default", "openbiz.css");
        // window lib
        BizSystem::clientProxy()->includePropWindowScripts();
        // validator lib
        //BizSystem::clientProxy()->includeValidatorScripts();
    }

    /**
     * Initialize all form objects.
     *
     * @return void
     */
    protected function initAllForms()
    {
        foreach ($this->m_FormRefs as $formRef)
        {
            $formName = $formRef->m_Name;
            $formObj = BizSystem::objectFactory()->getObject($formName);
            if ($formRef->m_SubForms && method_exists($formObj,"SetSubForms"))
                $formObj->setSubForms($formRef->m_SubForms);
        }
    }

    /**
     * Process request
     *
     * @return void
     */
    protected function processRequest()
    {
        // if url has form=...
        $paramForm = isset($_REQUEST['form']) ? $_REQUEST['form'] : null;
        // check url arg as fld:name=val
        $getKeys = array_keys($_REQUEST);
		$pageid = null;
		if(isset($_GET["pageid"])) {
			$pageid = $_GET["pageid"];
		}
        
        $paramFields = null;
        foreach ($getKeys as $key)
        {
            if (substr($key, 0, 4) == "fld:")
            {
                $fieldName = substr($key, 4);
                $fieldValue = $_REQUEST[$key];
                $paramFields[$fieldName] = $fieldValue;
            }
        }

        if (!$paramFields && !$pageid)
            return;

        // get the form object
        if (!$paramForm)
        { // get the first form name if no form is given
            foreach ($this->m_FormRefs as $formRef)
            {
                $paramForm = $formRef->m_Name;
                break;
            }
        }
        if (!$paramForm)
            return;
        $paramForm = $this->prefixPackage($paramForm);
        $formObj = BizSystem::objectFactory()->getObject($paramForm);
        $formObj->setRequestParams($paramFields);
        if($pageid){
			$formObj->setCurrentPage($pageid);
        }        
    }

    /**
     * Get output attributs
     * 
     * @return array
     * @todo need to raname to getOutputAttributs() or getAttributes
     */
    public function outputAttrs() 
    {
        $out['name'] = $this->m_Name;
        $out['description'] = $this->m_Description;
        $out["keywords"] = $this->m_Keywords;
        if ($this->m_Title)
            $title = Expression::evaluateExpression($this->m_Title,$this);
        else
        	$title = $this->m_Description;
        $out['title'] = $title;
        return $out;
    }
    
    protected function translate()
    {
    	$module = $this->getModuleName($this->m_Name);
    	$trans_string = I18n::t($this->m_Title, $this->getTransKey('Title'), $module);
    	if($trans_string){
    		$this->m_Title =  $trans_string;
    	}
    	$trans_string = I18n::t($this->m_Title, $this->getTransKey('Description'), $module);
    	if($trans_string){
    		$this->m_Description = $trans_string;
    	}
    }
    
    protected function getTransKey($name)
    {
    	$shortFormName = substr($this->m_Name,intval(strrpos($this->m_Name,'.'))+1);
    	return strtoupper($shortFormName.'_'.$name);
    }
}


/**
 * FormReference class is the class that contain form reference.
 *
 * @package openbiz.bin.easy
 * @author rocky swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class FormReference
{
    public $m_Name;
    public $m_SubForms;
    public $m_Description;
    private $_parentForm;
    public $m_Display = true;

    /**
     * Contructor, store form info from array to variable of class
     *
     * @param array $xmlArr array of form information
     */
    public function __construct($xmlArr)
    {
        $this->m_Name = $xmlArr["ATTRIBUTES"]["NAME"];
		if(isset($xmlArr["ATTRIBUTES"]["SUBFORMS"])) {
			$this->m_SubForms = $xmlArr["ATTRIBUTES"]["SUBFORMS"];
		}
		if(isset($xmlArr["ATTRIBUTES"]["DESCRIPTION"])) {
			$this->m_Description = $xmlArr["ATTRIBUTES"]["DESCRIPTION"];
		}
    }

    /**
     * Set parent form
     * 
     * @param string $formName form name
     * @@return void
     */
    public function setParentForm($formName)
    {
        $this->_parentForm = $formName;
    }
}

?>
