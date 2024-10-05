<?php
/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin.service
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

include_once (OPENBIZ_HOME."/messages/securityService.msg");

/**
 * securityService class is the plug-in service of security
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class securityService
{  
    public $m_Mode = 'DISABLED';
    private $_securityFilters = array();
    private $_messageFile;
    protected $m_ErrorMessage = null;

    /**
     * Initialize securityService with xml array metadata
     *
     * @param array $xmlArr
     * @return void
     */
    function __construct(&$xmlArr)
    {
        $this->readMetadata($xmlArr);
    }

    /**
     * Read array meta data, and store to meta object
     *
     * @param array $xmlArr
     * @return void
     */
    protected function readMetadata(&$xmlArr)
    {
        $this->m_Mode =   isset($xmlArr["PLUGINSERVICE"]["SECURITY"]["ATTRIBUTES"]["MODE"]) ? $xmlArr["PLUGINSERVICE"]["SECURITY"]["ATTRIBUTES"]["MODE"] : "DISABLED";
        if(strtoupper($this->m_Mode) == 'ENABLED' )
        {
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["URLFILTER"],		"securityFilter",	"URLFilter");
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["DOMAINFILTER"],	"securityFilter",	"DomainFilter");
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["IPFILTER"],		"securityFilter",	"IPFilter");
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["AGENTFILTER"],		"securityFilter",	"AgentFilter");
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["POSTFILTER"],		"securityFilter",	"PostFilter");
            $this->_securityFilters[] = new securityFilter($xmlArr["PLUGINSERVICE"]["SECURITY"]["GETFILTER"],		"securityFilter",	"GetFilter");
        }
    }

    /**
     * Get error message
     *
     * @return string
     */
    public function getErrorMessage()
    {
        return $this->m_ErrorMessage;
    }

    /**
     * Proses filter
     *
     * @return boolean
     */
    public function processFilters()
    {
        foreach($this->_securityFilters as $filter)
        {
            $filter->processRules();
            if($filter->getErrorMessage())
            {
                $this->m_ErrorMessage = $filter->getErrorMessage();
                return false;
            }
        }
        return true;

    }
}

/**
 * securityFilter class is helper class for security filter
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class securityFilter extends MetaIterator
{
    protected $m_Name = null;
    protected $m_Mode = 'DISABLED';
    protected $m_Rules = null;
    protected $m_ErrorMessage = null;
	protected $m_Package = null;


    /**
     * Initialize securityFilter with xml array metadata
     *
     * @param array $xmlArr
     * @param string $filterName
     * @param string $ruleName
     * @return void
     */
    function __construct(&$xmlArr, $filterName, $ruleName)
    {
        $this->readMetadata($xmlArr, $filterName, $ruleName);
    }

    /**
     * Read array meta data, and store to meta object
     *
     * @param array $xmlArr
     * @param string $filterName
     * @param string $ruleName
     * @return void
     */
    protected function readMetadata(&$xmlArr, $filterName, $ruleName)
    {
        $this->m_Name = $ruleName;
        $this->m_Mode =   isset($xmlArr["ATTRIBUTES"]["MODE"]) ? $xmlArr["ATTRIBUTES"]["MODE"] : "DISABLED";
        if(strtoupper($this->m_Mode) == 'ENABLED' )
        {
            $this->m_Rules 	= new MetaIterator($xmlArr["RULE"],	 $ruleName."Rule",	$this);
        }
    }

    /**
     * Get error message
     *
     * @return string
     */
    public function getErrorMessage()
    {
        return $this->m_ErrorMessage;
    }

    /**
     * Proses rule
     *
     * @return boolean|void
     */
    public function processRules()
    {
        if(isset($this->m_Rules->m_var) && is_array($this->m_Rules->m_var))
        {
            foreach($this->m_Rules->m_var as $name=>$obj)
            {
                $obj->process();
                if($obj->getErrorMessage())
                {
                    $this->m_ErrorMessage = $obj->getErrorMessage();
                    return false;
                }
            }
        }
    }
}

/**
 * iSecurityRule interface is interface for securityRule_Abstract
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
interface iSecurityRule
{
    /**
     * Proses security rule
     */
    public function process();
}

/**
 * securityRule_Abstract class is helper class for security filter
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class securityRule_Abstract implements iSecurityRule
{
    public $m_Name      =	null;
    public $m_Action    =	null;
    public $m_Match     =	null;
    public $m_Status     =	null;
    public $m_EffectiveTime =	null;
    public $m_ErrorMessage = null;

    /**
     * Initialize reportService with xml array metadata
     *
     * @param array $xmlArr
     * @return void
     */
    function __construct(&$xmlArr)
    {
        $this->readMetadata($xmlArr);
    }

    /**
     * Read array meta data, and store to meta object
     *
     * @param array $xmlArr
     * @return void
     */
    protected function readMetadata(&$xmlArr)
    {
        $this->m_Name 	= $xmlArr["ATTRIBUTES"]["NAME"];
        $this->m_Action	= $xmlArr["ATTRIBUTES"]["ACTION"];
        $this->m_Status	= $xmlArr["ATTRIBUTES"]["STATUS"];
        $this->m_Match 	= $xmlArr["ATTRIBUTES"]["MATCH"];
        $this->m_EffectiveTime = $xmlArr["ATTRIBUTES"]["EFFECTIVETIME"];
    }

    /**
     * Proses security rule
     *
     * @return string
     */
    public function process()
    {
        return true;
    }

    /**
     * Get message error
     *
     * @return string
     */
    public function getErrorMessage()
    {
        return $this->m_ErrorMessage;
    }

    /**
     * Check Effective Time
     *
     * @return boolean
     */
    public function checkEffectiveTime()
    {
        sscanf( $this->m_EffectiveTime, "%2d%2d-%2d%2d",
                $start_hour, $start_min,
                $end_hour, $end_min
        );

        $startTime  = strtotime(date("Y-m-d ").$start_hour.":".$start_min) ? strtotime(date("Y-m-d ").$start_hour.":".$start_min) : strtotime(date("Y-m-d 00:00"));
        $endTime    = strtotime(date("Y-m-d ").$end_hour.":".$end_min) ? strtotime(date("Y-m-d ").$end_hour.":".$end_min) : strtotime(date("Y-m-d 23:59:59"));

        $nowTime    = time();

        if($startTime>0 && $endTime>0)
        {
            //auto convert start time and end time
            if($endTime < $startTime)
            {
                $tmpTime = $startTime;
                $startTime = $endTime;
                $endTime = $tmpTime;
            }

            if($startTime < $nowTime && $nowTime < $endTime )
            {
                return true;
            }
            else
            {
                return false;
            }
        }
    }
}

/**
 * URLFilterRule class
 *
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class URLFilterRule extends securityRule_Abstract
{

    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
	        parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $url = $_SERVER['REQUEST_URI'];
	            if(preg_match("/".$this->m_Match."/si",$url))
	            {
	                if(strtoupper($this->m_Action)=='DENY')
	                {
	                    $this->m_ErrorMessage=BizSystem::getMessage('SECURITYSVC_URL_DENIED');
	                    return false;
	                }elseif(strtoupper($this->m_Action)=='ALLOW')
	                {
	                    return true;
	                }
	                return false;
	            }
	        }
    	}
    }
}

/**
 * DomainFilterRule class
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class DomainFilterRule extends securityRule_Abstract
{

    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
	        parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $url = $_SERVER['HTTP_HOST'];
	            if(preg_match("/".$this->m_Match."/si",$url))
	            {
	                if(strtoupper($this->m_Action)=='DENY')
	                {
	                    $this->m_ErrorMessage=BizSystem::getMessage('SECURITYSVC_DOMAIN_DENIED');
	                    return false;
	                }
	                elseif(strtoupper($this->m_Action)=='ALLOW')
	                {
	                    return true;
	                }
	                return false;
	            }
	        }
    	}
    }
}

/**
 * AgentFilterRule class
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class AgentFilterRule extends securityRule_Abstract
{
    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
	        parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $url = $_SERVER['HTTP_USER_AGENT'];
	            if(preg_match("/".$this->m_Match."/si",$url))
	            {
	                if(strtoupper($this->m_Action)=='DENY')
	                {
	                    $this->m_ErrorMessage=BizSystem::getMessage('SECURITYSVC_AGENT_DENIED');
	                    return false;
	                }
	                elseif(strtoupper($this->m_Action)=='ALLOW')
	                {
	                    return true;
	                }
	                return false;
	            }
	        }
    	}
    }
}

/**
 * IPFilterRule class
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class IPFilterRule extends securityRule_Abstract
{
    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
    		parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $url = $_SERVER['REMOTE_ADDR'];
	            if(preg_match("/".$this->m_Match."/si",$url))
	            {
	                if(strtoupper($this->m_Action)=='DENY')
	                {
	                    $this->m_ErrorMessage = BizSystem::getMessage('SECURITYSVC_IPADDR_DENIED');
	                    return false;
	                }
	                elseif(strtoupper($this->m_Action)=='ALLOW')
	                {
	                    return true;
	                }
	                return false;
	            }
	        }
    	}
    }
}

/**
 * PostFilterRule class
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class PostFilterRule extends securityRule_Abstract
{

    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
	        parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $post_str = serialize($_POST);
	            if($this->m_Match!="")
	            {
	                if(preg_match("/".$this->m_Match."/si",$post_str))
	                {
	                    if(strtoupper($this->m_Action)=='DENY')
	                    {
	                        $this->m_ErrorMessage=BizSystem::getMessage('SECURITYSVC_POST_DENIED');
	                        return false;
	                    }
	                    elseif(strtoupper($this->m_Action)=='ALLOW')
	                    {
	                        return true;
	                    }
	                    return false;
	                }
	            }
	            else
	            {
	                return false;
	            }
	        }
    	}
    }
}

/**
 * GetFilterRule class
 *
 * @package   openbiz.bin.service
 * @author    Rocky Swen
 * @copyright Copyright (c) 2003-2009, Rocky Swen
 * @access    public
 */
class GetFilterRule extends securityRule_Abstract
{

    /**
     * Proses security rule
     * return true go to check next rule
     * return false report an error and stop checking
     *
     * @return boolean
     */
    public function process()
    {
    	if(strtoupper($this->m_Status)=='ENABLE')
    	{
	        parent::process();
	        if(!$this->checkEffectiveTime())
	        {
	            return true;
	        }
	        else
	        {
	            $get_str = serialize($_GET);
	            if(preg_match("/".$this->m_Match."/si",$get_str))
	            {
	                if(strtoupper($this->m_Action)=='DENY')
	                {
	                    $this->m_ErrorMessage=BizSystem::getMessage('SECURITYSVC_GET_DENIED');
	                    return false;
	                }
	                elseif(strtoupper($this->m_Action)=='ALLOW')
	                {
	                    return true;
	                }
	                return false;
	            }
	        }
    	}
    }
}
?>
