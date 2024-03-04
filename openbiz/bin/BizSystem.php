<?PHP
/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

// register __destruct method as shutdown function
register_shutdown_function("bizsystem_shutdown");

function bizsystem_shutdown()
{
    BizSystem::sessionContext()->saveSessionObjects();
}

include_once 'Resource.php';

/**
 * BizSystem class
 *
 * BizSystem is initialized for each request, it provides infrastructure objects
 * and utility methods which are used in whole request.
 * BizSystem is singleton pattern class that can create instant
 * with BizSystem::instance.
 *
 * @package   openbiz.bin
 * @author    Rocky Swen
 * @copyright Copyright (c) 2005-2009, Rocky Swen
 * @access    public
 */
class BizSystem
{
    private $_objectFactory = null;
    private $_sessionContext = null; // instant of SessionContext class
    private $_confgiuration = null;
    private $_clientProxy = null;
    private $_typeManager = null;
    private $_currentViewName = "";
    private $_currentViewSet = "";
    private $_dbConnection = array();
    private $_theme = array();       // TODO: unused variable
    private $_serviceList = array(); // TODO: unused variable
    private $_userProfile = null;    // TODO: unused variable (store on Session with SessionContext)

    private static $_instance = null;

    /**
     * Create instant of BizSystem
     *
     * @return BizSystem instant of BizSystem
     */
    public static function instance()
    {
        if (self::$_instance == null)
            self::$_instance = new BizSystem();
        return self::$_instance;
    }

    /**
     * Construct object: initialize SessionContext and retieve object session variables
     *
     * @return void
     */
    private function __construct()
    {
        include_once(OPENBIZ_BIN."SessionContext.php");
        $this->_sessionContext = new SessionContext();
        // retrieve object session vars
        $this->_sessionContext->retrieveSessionObjects();
    }

    /**
     * Destruct object: save object session variables
     *
     * @return void
     */
    public function __destruct()
    {
        // save object session vars
        //$this->_sessionContext->saveSessionObjects();
        //echo "<br>destruct bizSystem";
    }

    /**
     * Return the version of phpOpenBiz
     *
     * @return string
     */
    public static function getVersion()
    {
        return '2.4beta';
    }

    /**
     * Get the ObjectFactory object
     *
     * @return ObjectFactory the ObjectFactory object
     */
    public function getObjectFactory()
    {
        if (!$this->_objectFactory)
        {
            include_once(OPENBIZ_BIN."ObjectFactory.php");
            $this->_objectFactory = new ObjectFactory();
        }
        return $this->_objectFactory;
    }

    /**
     * Get get the ObjectFactory object as static method
     * this static method wrapping the GetObjectFactory method
     *
     * @return ObjectFactory the ObjectFactory object
     */
    public static function objectFactory()
    {
        return BizSystem::instance()->getObjectFactory();
    }

    /**
     * Get the SessionContext object
     *
     * @return SessionContext the SessionContext object
     */
    public function getSessionContext()
    {
        return $this->_sessionContext;
    }


    /**
     * Get the SessionContext object
     * this static method wrapping the GetObjectFactory method
     *
     * @return SessionContext the SessionContext object
     */
    public static function sessionContext()
    {
        return BizSystem::instance()->getSessionContext();
    }

    /**
     * Get the Configuration object
     *
     * @return Configuration the Configuration object
     */
    public function getConfiguration()
    {
        if (!$this->_confgiuration)
        {
            include_once(OPENBIZ_BIN."Configuration.php");
            $this->_confgiuration = new Configuration();
        }
        return $this->_confgiuration;
    }

    /**
     * Get the Configuration object
     * this static method wrapping the Configuration method
     *
     * @return Configuration the Configuration object
     */
    public static function configuration()
    {
        return BizSystem::instance()->getConfiguration();
    }

    /**
     * Get the ClientProxy object
     *
     * @return ClientProxy the ClientProxy object
     */
    public function getClientProxy()
    {
        if (!$this->_clientProxy)
        {
            include_once(OPENBIZ_BIN."ClientProxy.php");
            $this->_clientProxy = new ClientProxy();
        }
        return $this->_clientProxy;
    }

    /**
     * Get the ClientProxy object
     * this static method wrapping the GetClientProxy method
     *
     * @return ClientProxy the ClientProxy object
     */
    public static function clientProxy()
    {
        return BizSystem::instance()->getClientProxy();
    }

    /**
     * Get the TypeManager object
     *
     * @return TypeManager the TypeManager object
     */
    public function getTypeManager()
    {
        if (!$this->_typeManager)
        {
            include_once(OPENBIZ_BIN."TypeManager.php");
            $this->_typeManager = new TypeManager();
        }
        return $this->_typeManager;
    }

    /**
     * Get the TypeManager object
     * this static method wrapping the GetTypeManager method
     *
     * @return TypeManager the TypeManager object
     */
    public static function typeManager()
    {
        return BizSystem::instance()->getTypeManager();
    }

    /**
     * Get the service object
     *
     * @param string $service service name
     * @return object the service object
     */
    public static function getService($service, $new=0)
    {
        $default_package = "service";
        $svc_name = $service;
        if (strpos($service, ".") === false)
            $svc_name = $default_package . "." . $service;
        return BizSystem::getObject($svc_name, $new);
    }
    
    /**
     * Get the metadata object
     *
     * @param string $objectName object name
     * @return <object> the object
     */
    public static function getObject($objectName, $new=0)
    {
        return BizSystem::ObjectFactory()->getObject($objectName, $new);
    }

    /**
     * Check if user can access the given resource action
     *
     * @param string $resourceAction resource action
     * @return boolean true or false
     */
    public static function allowUserAccess($resourceAction)
    {
        $serviceObj = BizSystem::getService(ACL_SERVICE);
        return $serviceObj->allowAccess($resourceAction);
    }

    /**
     * Initialize User Profile
     *
     * @param string $userId
     * @return array Profile array
     */
    public static function initUserProfile($userId)
    {
        $serviceObj = BizSystem::getService(PROFILE_SERVICE);

        if (method_exists($serviceObj,'InitProfile'))
            $profile = $serviceObj->InitProfile($userId);
        else
            $profile = $serviceObj->getProfile($userId);

        BizSystem::sessionContext()->setVar("_USER_PROFILE", $profile);
        return $profile;
    }

    /**
     * Get user profile
     *
     * @param string $attribute user attribute
     * @return array user profile array
     */
    public static function getUserProfile($attribute=null)
    {
        $serviceObj = BizSystem::getService(PROFILE_SERVICE);
        if (method_exists($serviceObj,'getProfile'))
            return $serviceObj->getProfile($attribute);
        else
        {
            $profile = BizSystem::sessionContext()->getVar("_USER_PROFILE");
            return isset($profile[$attribute]) ? $profile[$attribute] : "";
        }
    }

    public static function getProfileName($account_id){
    	$serviceObj = BizSystem::getService(PROFILE_SERVICE);
    	return $serviceObj->GetProfileName($account_id);
    }
    /**
     * Get the current view name
     *
     * @return string current view name
     */
    public function getCurrentViewName()
    {
        if ($this->_currentViewName == "")
            $this->_currentViewName = $this->getSessionContext()->getVar("CVN");   // CVN stands for CurrentViewName
        return $this->_currentViewName;
    }

    /**
     * Set the current view name
     *
     * @param string $viewname new current view name
     */
    public function setCurrentViewName($viewname)
    {
        $this->_currentViewName = $viewname;
        $this->getSessionContext()->setVar("CVN", $this->_currentViewName);   // CVN stands for CurrentViewName
    }

    /**
     * Get the current view set
     *
     * @return string current view set
     */
    public function getCurrentViewSet()
    {
        if ($this->_currentViewSet == "")
            $this->_currentViewSet = $this->getSessionContext()->getVar("CVS");   // CVS stands for CurrentViewSet
        return $this->_currentViewSet;
    }

    /**
     * Set current view set
     *
     * @param <type> $viewSet
     */
    public function setCurrentViewSet($viewSet)
    {
        $this->_currentViewSet = $viewSet;
        $this->getSessionContext()->setVar("CVS", $this->_currentViewSet);   // CVS stands for CurrentViewSet
    }

    /**
     * Get current page URL
     * NOTE: NYU not yet used
     *
     * @return string current page URL
     */
    public static function currentPageURL()
    {
        if ($_REQUEST['__url'])
            return $_REQUEST['__url'];
        else
            return $_SERVER['REQUEST_URI'];
    }

    /**
     * Get the database connection object
     *
     * @param string $dbname, database name declared in config.xml
     * @return Zend_Db_Adapter_Abstract
     */
    public function getDBConnection($dbName=null)
    {
        $rDBName = (!$dbName) ? "Default" : $dbName;
        if (isset($this->_dbConnection[$rDBName]))
            return $this->_dbConnection[$rDBName];

        $dbInfo = $this->getConfiguration()->getDatabaseInfo($rDBName);

        require_once 'Zend/Db.php';

        $params = array (
                'host'     => $dbInfo["Server"],
                'username' => $dbInfo["User"],
                'password' => $dbInfo["Password"],
                'dbname'   => $dbInfo["DBName"],
                'port'     => $dbInfo["Port"],
                'charset'  => $dbInfo["Charset"]
        );
        if ($dbInfo["Options"]) {
        	$options = explode(";",$dbInfo["Options"]);
	        foreach ($options as $opt) {
	        	list($k,$v) = explode("=",$opt);
	        	$params[$k] = $v;
	        }
        }
        foreach ($params as $name=>$val) {
        	if (empty($val)) unset($params[$name]);
        }
        if(strtoupper($dbInfo["Driver"])=="PDO_MYSQL")
        {
        	$pdoParams = array(
    			PDO::MYSQL_ATTR_USE_BUFFERED_QUERY => true
			);
			$params["driver_options"]=$pdoParams;
        }
        $db = Zend_Db::factory($dbInfo["Driver"], $params);

        $db->setFetchMode(PDO::FETCH_NUM);
        
        if(strtoupper($dbInfo["Driver"])=="PDO_MYSQL" &&
                $dbInfo["Charset"]!="")
        {
            $db->query("SET NAMES '".$params['charset']."'");
        }
        
        $this->_dbConnection[$rDBName] = $db;

        return $db;
    }

    /**
     * Get database connection
     * 
     * @param string $dbName database name
     * @return Zend_Db_Adapter_Abstract database connection
     */
    public static function dbConnection($dbName=null)
    {
        return BizSystem::instance()->getDBConnection($dbName);
    }

    /**
     * get identifier quoted as database table if needed
     * @retutn string
     */
    public static function getDBTableQuoted($identifier, $dbName=null)
	{
		$rDBName = (!$dbName) ? "Default" : $dbName;

        $dbInfo = BizSystem::instance()->getConfiguration()->getDatabaseInfo($rDBName);
		if(strtoupper($dbInfo["Driver"])=="PDO_MYSQL") {
			$value = trim(trim($identifier), "`");
			return $value;
		}
		return $identifier;
    }

    /**
     * Evaluate macro, this method can only be used to get profile in 2.0
     * For example, @macro_var:macro_key. i.e. @profile:ROLE
     *
     * @param string $var, macro name
     * @param string $key, macro key
     * @return string
     */
    public static function getMacroValue($var, $key)
    {
        if ($var == "profile")
        {
            return BizSystem::instance()->getUserProfile($key);
        }
        return null;
    }

    /**
     * Get smarty template
     *
     * @return Smarty smarty object
     */
    public static function getSmartyTemplate()
    {
        return Resource::getSmartyTemplate();
    }

    /**
     * Get Zend Template
     *
     * @return Zend_View zend view template object
     */
    public static function getZendTemplate()
    {
        return Resource::getZendTemplate();
    }

    /**
     * Log message to log file
     *
     * @param integer $priority. it can be one of following value
     *    LOG_EMERG	system is unusable = 1
     *    LOG_ALERT	action must be taken immediately = LOG_EMERG
     *    LOG_CRIT	   critical conditions = LOG_EMERG
     *    LOG_ERR	   error conditions = 4
     *    LOG_WARNING	warning conditions = 5
     *    LOG_NOTICE	normal, but significant, condition = 6
     *    LOG_INFO	   informational message = LOG_NOTICE
     *    LOG_DEBUG	debug-level message = LOG_NOTICE
     *    ### So LOG_EMERG, LOG_ERR, LOG_WARNING and LOG_DEBUG are valid inputs ###
     * @param string $subject. the log subject decided by caller function
     * @param string $message. the message to be logged in log file
     * @return void
     */
    public static function log($priority, $subject, $message)
    {
        $svcobj = BizSystem::getService(LOG_SERVICE);
        $svcobj->log($priority, $subject, $message);
    }

    /**
     * Log message to log file
     *
     * @global BizSystem $g_BizSystem
     * @param integer $priority. it can be one of following value
     *    LOG_EMERG	system is unusable = 1
     *    LOG_ALERT	action must be taken immediately = LOG_EMERG
     *    LOG_CRIT	   critical conditions = LOG_EMERG
     *    LOG_ERR	   error conditions = 4
     *    LOG_WARNING	warning conditions = 5
     *    LOG_NOTICE	normal, but significant, condition = 6
     *    LOG_INFO	   informational message = LOG_NOTICE
     *    LOG_DEBUG	debug-level message = LOG_NOTICE
     *    ### So LOG_EMERG, LOG_ERR, LOG_WARNING and LOG_DEBUG are valid inputs ###
     *
     * @param string $subject. the log subject decided by caller function
     * @param string $message. the message to be logged in log file
     * @param string $fileName file to save to
     * @return void
     */
    public static function logError($priority, $subject, $message, $fileName = NULL)
    {
        $svcobj = BizSystem::getService(LOG_SERVICE);
        $svcobj->logError($priority, $subject, $message, $fileName);
    }

    /**
     * Get Xml file with path
     *
     * Search the object metedata file as objname+.xml in metedata directories
     * name convension: demo.BOEvent points to metadata/demo/BOEvent.xml
     * new in 2.2.3, demo.BOEvent can point to modules/demo/BOEvent.xml
     *
     * @param string $xmlObj xml object
     * @return string xml config file path
     **/
    public static function getXmlFileWithPath($xmlObj)
    {
        return Resource::GetXmlFileWithPath($xmlObj);
    }

    /**
     * Get openbiz template file path by searching modules/package, /templates
     *
     * @param string $className
     * @return string php library file path
     **/
    public static function getTplFileWithPath($templateFile, $packageName)
    {
        return Resource::getTplFileWithPath($templateFile, $packageName);
    }

    /**
     * Get openbiz library php file path by searching modules/package, /bin/package and /bin
     *
     * @param string $className
     * @return string php library file path
     **/
    public static function getLibFileWithPath($className, $packageName="")
    {
        return Resource::getLibFileWithPath($className, $packageName);
    }

    /**
     * Get core path of class
     * NOTE: unused
     *
     * @param string $className class name
     * @return string full file name of class
     */
    private static function _getCoreLibFilePath($className)
    {
        return Resource::getCoreLibFilePath($className);
    }

    /**
     * Get Xml Array.
     * If xml file has been compiled (has .cmp), load the cmp file as array;
     * otherwise, compile the .xml to .cmp first new 2.2.3, .cmp files
     * will be created in app/cache/metadata_cmp directory. replace '/' with '_'
     * for example, /module/demo/BOEvent.xml has cmp file as _module_demo_BOEvent.xml
     *
     * @param string $xmlFile
     * @return array
     **/
    public static function &getXmlArray($xmlFile)
    {
        return Resource::getXmlArray($xmlFile);
    }

    /**
     * Get message resource
     * enhanced by Jixian
     *
     * @param <type> $msgid  string id of message
     * @param <type> $params array array of parameters
     * @return string message
     */
    public static function getMessage($msgid, $params=array())
    {
        return Resource::getMessage($msgid, $params);
    }

}

?>
