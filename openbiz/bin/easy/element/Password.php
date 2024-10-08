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
include_once ('InputElement.php');

/**
 * Password class is element for input password
 *
 * @package openbiz.bin.easy.element
 * @author Rocky Swen
 * @copyright Copyright (c) 2005-2009
 * @access public
 */
class Password extends InputElement
{
    public function readMetaData(&$xmlArr)
    {
        parent::readMetaData($xmlArr);
        $this->m_cssClass = isset($xmlArr['ATTRIBUTES']['CSSCLASS']) ? $xmlArr['ATTRIBUTES']['CSSCLASS'] : 'input_text';
        $this->m_cssErrorClass = isset($xmlArr['ATTRIBUTES']['CSSERRORCLASS']) ? $xmlArr['ATTRIBUTES']['CSSERRORCLASS'] : $this->m_cssClass . '_error';
        $this->m_cssFocusClass = isset($xmlArr['ATTRIBUTES']['CSSFOCUSCLASS']) ? $xmlArr['ATTRIBUTES']['CSSFOCUSCLASS'] : $this->m_cssClass . '_focus';
    }

    /**
     * Render, draw the control according to the mode
     *
     * @return string HTML text
     */
    public function render()
    {
        $disabledStr = ($this->getEnabled() == 'N') ? 'DISABLED="true"' : '';
        $style = $this->getStyle();
        $formobj = $this->GetFormObj();
        $func = '';
        if (isset($formobj->m_Errors) && isset($formobj->m_Errors[$this->m_Name])) {
            $func .= "onchange=\"this.className='$this->m_cssClass'\"";
        } else {
            $func .= "onfocus=\"this.className='$this->m_cssFocusClass'\" onblur=\"this.className='$this->m_cssClass'\"";
        }
        $sHTML = "<INPUT TYPE=\"PASSWORD\" NAME='$this->m_Name' ID=\"" . $this->m_Name . "\" VALUE='$this->m_Value' $disabledStr $this->m_HTMLAttr $style $func />";

        return $sHTML;
    }
}

?>
