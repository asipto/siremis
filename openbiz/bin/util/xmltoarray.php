<?php

/**
 * PHPOpenBiz Framework
 *
 * LICENSE
 *
 * This source file is subject to the BSD license that is bundled
 * with this package in the file LICENSE.txt.
 *
 * @package   openbiz.bin.util
 * @copyright Copyright &copy; 2005-2009, Rocky Swen
 * @license   http://www.opensource.org/licenses/bsd-license.php
 * @link      http://www.phpopenbiz.org/
 * @version   $Id$
 */

// Based on code found online at:
// http://php.net/manual/en/function.xml-parse-into-struct.php
//
// Author: Eric Pollmann
// Released into public domain September 2003
// http://eric.pollmann.net/work/public_domain/

/*
 * //$parser = new XMLParser('./BOEvent.xml', 'file', 1);
 * //$parser = new XMLParser('./FMSponsor_chart.xml', 'file', 1);
 * //$parser = new XMLParser('./EventView.xml', 'file', 1);
 * $parser = new XMLParser('./FMEvent.xml', 'file', 1);
 * $tree = $parser->getTree();
 * echo "<pre>";
 * print_r($tree);
 * echo "</pre>";
 */
class XMLParserX
{
    var $data;  // Input XML data buffer
    var $vals;  // Struct created by xml_parse_into_struct
    var $collapse_dups;  // If there is only one tag of a given name,
    //   shall we store as scalar or array?

    var $index_numeric;  // Index tags by numeric position, not name.
    //   useful for ordered XML like CallXML.

    // Read in XML on object creation.
    // We can take raw XML data, a stream, a filename, or a url.
    function __construct($data_source, $data_source_type = 'raw', $collapse_dups = 0, $index_numeric = 0)
    {
        $this->collapse_dups = $collapse_dups;
        $this->index_numeric = $index_numeric;
        $this->data = '';
        if ($data_source_type == 'raw') {
            $this->data = $data_source;
        } elseif ($data_source_type == 'stream') {
            while (!feof($data_source)) {
                $this->data .= fread($data_source, 1000);
            }
            // try filename, then if that fails...
        } elseif (file_exists($data_source)) {
            $this->data = implode('', file($data_source));
        } else {
            // try url
            $fp = fopen($data_source, 'r');
            while (!feof($fp)) {
                $this->data .= fread($fp, 1000);
            }
            fclose($fp);
        }
    }

    function XMLParserX($data_source, $data_source_type = 'raw', $collapse_dups = 0, $index_numeric = 0)
    {
        self::__construct($data_source, $data_source_type, $collapse_dups, $index_numeric);
    }

    // Parse the XML file into a verbose, flat array struct.
    // Then, coerce that into a simple nested array.
    function &getTree()
    {
        // load xml and check if it turns a valid object. TODO: it will throw error.
        /*$xml = simplexml_load_file($this->data);
        if (!$xml) {
            echo 'invalid xml file '.$this->data; exit;
        }*/
        // $parser = xml_parser_create('ISO-8859-1');
        $parser = xml_parser_create('');
        xml_parser_set_option($parser, XML_OPTION_SKIP_WHITE, 1);
        xml_parse_into_struct($parser, $this->data, $vals, $index);
        xml_parser_free($parser);
        // print_r($vals);
        $results = print_r($vals, true);

        $i = -1;
        $clist = $this->getchildren($vals, $i);
        return $clist;
    }

    // internal function: build a node of the tree
    function buildtag($thisvals, $vals, $i, $type)
    {
        if (isset($thisvals['attributes'])) {
            $tag['ATTRIBUTES'] = $thisvals['attributes'];
        } else {
            $tag = array();
        }
        // complete tag, just return it for storage in array
        if ($type === 'complete' && isset($thisvals['value'])) {
            $tag['VALUE'] = $thisvals['value'];
        } else {
            // open tag, recurse
            $tag = array_merge((array) $tag, (array) $this->getchildren($vals, $i));
        }
        return $tag;
    }

    // internal function: build an nested array representing children
    function getchildren($vals, &$i)
    {
        $children = array();  // Contains node data

        // print_r($vals);

        // Node has CDATA before it's children
        if ($i > -1 && isset($vals[$i]['value'])) {
            $children['VALUE'] = $vals[$i]['value'];
        }
        $c = count($vals);
        // Loop through children, until hit close tag or run out of tags
        while (++$i < $c) {
            $type = $vals[$i]['type'];

            // 'cdata':	Node has CDATA after one of it's children
            // 		(Add to cdata found before in this case)
            if ($type === 'cdata') {
                $children['VALUE'] .= $vals[$i]['value'];
            } elseif ($type === 'complete' || $type === 'open') {
                // 'complete':	At end of current branch
                // 'open':	Node has children, recurse
                $tag = $this->buildtag($vals[$i], $vals, $i, $type);
                if ($this->index_numeric) {
                    $tag['TAG'] = $vals[$i]['tag'];
                    $children[] = $tag;
                } else {
                    if (isset($vals[$i]) && isset($vals[$i]['tag'])) {
                        $children[$vals[$i]['tag']][] = $tag;
                    }
                }
            } elseif ($type === 'close') {
                // 'close:	End of node, return collected data
                //		Do not increment $i or nodes disappear!
                break;
            }
        }
        if ($this->collapse_dups) {
            foreach ($children as $key => $value) {
                if (is_array($value) && (count($value) == 1)) {
                    $children[$key] = $value[0];
                }
            }
        }
        return $children;
    }
}

?>
