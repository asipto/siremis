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

/**
 * ObjectFactory is factory class to create metadata based objects
 * (bizview, bizform, bizdataobj...)
 *
 * @package   openbiz.bin
 * @author    Rocky Swen <rocky@phpopenbiz.org>
 * @copyright Copyright (c) 2005-2009, Rocky Swen
 * @access    public
 */
class ObjectFactory
{
    /**
     * Internal array for cache MetaObject
     * @var array
     */
    protected $_objsRefMap = array();

    public function __construct()
    {
    }

    public function __destruct()
    {
    }

    /**
     * Get a metadata based object instance.
     * It returns the instance the internal object map or create a new one and save it in the map.
     *
     * @param string $objName name of object that want to get
     * @return object
     */
    public function getObject($objName, $new=0)
    {
        if (array_key_exists($objName, $this->_objsRefMap) && $new==0)
        {
            return $this->_objsRefMap[$objName];
        }

        $obj = $this->constructObject($objName);
        if ($obj)
            $this->_objsRefMap[$objName] = $obj; // save object to cache
        if (method_exists($obj, "GetSessionVars"))
            $obj->getSessionVars(BizSystem::sessionContext());

        return $obj;
    }

    /**
     * Create a new metadata based object instance
     *
     * @param string $objName name of object will be create
     * @param array $xmlArr propery array of object
     * @return object
     */
    public function createObject($objName, &$xmlArr=null)
    {
        $obj = $this->constructObject($objName, $xmlArr);
        return $obj;
    }

    public function setObject($objName, $obj)
    {
        $this->_objsRefMap[$objName] = $obj;
    }

    /**
     * Get all object from the internal object array (object cache)
     *
     * @return array array of object
     */
    public function getAllObjects()
    {
        return $this->_objsRefMap;
    }

    /**
     * Construct an instance of an object
     *
     * @param string $objName object name
     * @param array $xmlArr xml array
     * @return object the instance of the object
     */
    protected function constructObject($objName, &$xmlArr=null)
    {
        if (!$xmlArr || (count($xmlArr) == 0))
        {
            $xmlFile = BizSystem::GetXmlFileWithPath ($objName);
            if (!$xmlFile)
            {
                $dotPos = strrpos($objName, ".");
                $package = $dotPos>0 ? substr($objName, 0, $dotPos) : null;
                $class = $dotPos>0 ? substr($objName, $dotPos+1) : $objName;
            }
            else {
                $xmlArr = BizSystem::getXmlArray($xmlFile);
            }
        }
        if ($xmlArr)
        {
            $keys = array_keys($xmlArr);
            $root = $keys[0];

            // add by mr_a_ton , atrubut name must match with object name
            $dotPos = strrpos($objName, ".");
            $shortObjectName  = $dotPos > 0 ? substr($objName, $dotPos+1) : $objName;
            if ($xmlArr[$root]["ATTRIBUTES"]["NAME"]=="")
            {
                $xmlArr[$root]["ATTRIBUTES"]["NAME"]=$shortObjectName;
            }
            else
            {
                if ($shortObjectName != $xmlArr[$root]["ATTRIBUTES"]["NAME"] )
                {
                    trigger_error("Metadata file parsing error for object $objName. Name attribut [".$xmlArr[$root]["ATTRIBUTES"]["NAME"]."] not same with object name. Please double check your metadata xml file again.", E_USER_ERROR);
                }
            }

            //$package = $xmlArr[$root]["ATTRIBUTES"]["PACKAGE"];
            $class = $xmlArr[$root]["ATTRIBUTES"]["CLASS"];
            // if class has package name as prefix, change the package to the prefix
            $dotPos = strrpos($class, ".");
            $classPrefix = $dotPos>0 ? substr($class, 0, $dotPos) : null;
            $classPackage = $classPrefix ? $classPrefix : null;
            if ($classPrefix) $class = substr($class, $dotPos+1);
            // set object package
            $dotPos = strrpos($objName, ".");
            $package = $dotPos>0 ? substr($objName, 0, $dotPos) : null;
            if (!$classPackage) $classPackage = $package;
            $xmlArr[$root]["ATTRIBUTES"]["PACKAGE"] = $package;
	}

        if ($class == "BizObj")  // convert BizObj to BizDataObj, support <1.2 version
            $class = "BizDataObj";

        if (!class_exists($class, false))
        {
            $classFile= BizSystem::getLibFileWithPath($class, $classPackage);
            if (!$classFile)
            {
                if ($package)
                    trigger_error("Cannot find the class with name as '$package.$class'", E_USER_ERROR);
                else
                    trigger_error("Cannot find the class with name as '$class'", E_USER_ERROR);
                exit();
            }
            include_once($classFile);
        }
        if (class_exists($class, false))
        {
            $obj_ref = new $class($xmlArr);
            if ($obj_ref)
            {
                return $obj_ref;
            }
        }
        else
            trigger_error("Cannot find the class with name as $class in $classFile", E_USER_ERROR);
        return null;
    }

}

?>
