<?php
// map url parameters to openbiz view, form, ...
// http://localhost/?/user/login			 => http://localhost/bin/controller.php?view=user.view.LoginView
// http://localhost/?/user/reset_password => http://localhost/bin/controller.php?view=user.view.RestPasswordView
// http://localhost/?/article/1 			 => http://localhost/bin/controller.php?view=page.view.ArticleView&fld:Id=1
// ($DEFAULT_MODULE="page")
// http://localhost/?/article/1/f_catid_20=> http://localhost/bin/controller.php?view=page.view.ArticleView&fld:Id=1&fld:catid=20
// ($DEFAULT_MODULE="page")
// http://localhost/?/article/catid_20 	 => http://localhost/bin/controller.php?view=page.view.ArticleView&catid=20
// ($DEFAULT_MODULE="page")
include 'app.inc';

$DEFAULT_VIEW = 'LoginView';
$DEFAULT_MODULE = 'user';
$DEFAULT_URL = 'index.php/user/login';

if (isset($_SERVER['REDIRECT_QUERY_STRING'])) {
    $url = $_SERVER['REDIRECT_QUERY_STRING'];
} elseif (isset($_SERVER['REQUEST_URI']) && preg_match('/\?\/?(.*?)(\.html)?$/si', $_SERVER['REQUEST_URI'], $match)) {
    // supports for http://localhost/?/user/login format
    // supports for http://localhost/index.php?/user/login format
    $url = $match[1];
} elseif (strlen($_SERVER['REQUEST_URI']) > strlen($_SERVER['SCRIPT_NAME'])) {
    // supports for http://localhost/index.php/user/login format
    $url = str_replace($_SERVER['SCRIPT_NAME'], '', $_SERVER['REQUEST_URI']);
    preg_match('/\/?(.*?)(\.html)?$/si', $url, $match);
    $url = $match[1];
} else {
    // REQUEST_URI = /cubi/
    // SCRIPT_NAME = /cubi/index.php
    $url = '';
}

// remove repeat slash //
$url = preg_replace('/([\/\/]+)/', '/', $url);
preg_match('/\/?(.*?)(\.html)?$/si', $url, $match);
$url = $match[1];

$urlArr = array();
if ($url) {
    $profile = BizSystem::getUserProfile();
    if (!$profile) {
        if ('index.php/' . $url != $DEFAULT_URL) {
            header('Location: ' . APP_INDEX . '/' . $DEFAULT_URL);
        }
        $module_name = $DEFAULT_MODULE;
        $view_name = $DEFAULT_VIEW;
    } else {
        $urlArr = preg_split('/\//si', $url);
        if (preg_match('/^[a-z_]*$/si', $urlArr[1])) {
            // http://localhost/?/ModuleName/FormName/
            $module_name = $urlArr[0];
            $view_name = getViewName($urlArr[1]);
        } elseif (preg_match('/^[a-z_]*$/si', $urlArr[0])) {
            // http://localhost/?/FormName/
            $module_name = $DEFAULT_MODULE;
            $view_name = getViewName($urlArr[0]);
        }
        if (empty($urlArr[count($urlArr) - 1])) {
            unset($urlArr[count($urlArr) - 1]);
        }
    }
} else {
    // http://localhost/
    $module_name = $DEFAULT_MODULE;
    $view_name = $DEFAULT_VIEW;
    $profile = BizSystem::getUserProfile();
    if (isset($profile['roleStartpage']) && isset($profile['roleStartpage'][0])) {
        $DEFAULT_URL = APP_INDEX . $profile['roleStartpage'][0];
    }
    header("Location: $DEFAULT_URL");
}

$TARGET_VIEW = $module_name . '.view.' . $view_name;
$_GET['view'] = $_REQUEST['view'] = $TARGET_VIEW;

$PARAM_MAPPING = getParameters($urlArr);
if (isset($PARAM_MAPPING)) {
    foreach ($PARAM_MAPPING as $param => $value) {
        // if (isset($_GET[$param]))
        $_GET[$param] = $_REQUEST[$param] = $value;
    }
}

include dirname(__FILE__) . '/controller.php';

function getViewName($url_path)
{
    if (preg_match_all('/([a-z]*)_?/si', $url_path, $match)) {
        $view_name = '';
        $match = $match[1];
        foreach ($match as $part) {
            if ($part) {
                $part = ucwords($part);  // ucwords(strtolower($part));
                $view_name .= $part;
            }
        }
        $view_name .= 'View';
    }
    return $view_name;
}

function getParameters($urlArr)
{
    $PARAM_MAPPING = array();
    // foreach($urlArr as $path)
    for ($i = 2; $i < count($urlArr); $i++)  // ignore the first 2 parts
    {
        // only numberic like 20 parse it as fld:Id=20
        if (preg_match('/^([0-9]*)$/si', $urlArr[$i], $match)) {
            $PARAM_MAPPING['fld:Id'] = $match[1];
            continue;
        }
        // Cid_20 parse it as fld:Cid=20
        // http://localhost/cubi/some/thing/Cid_20
        // echo $_GET['Cid'];  // 20
        elseif (preg_match('/^([a-z]*?)_([a-z0-9]*)$/si', $urlArr[$i], $match)) {
            $PARAM_MAPPING['fld:' . $match[1]] = $match[2];
            $_GET[$match[1]] = $match[2];
            continue;
        }
        // parse the string to query string
        parse_str($urlArr[$i], $arr);
        foreach ($arr as $k => $v) {
            $_GET[$k] = $v;
            $PARAM_MAPPING[$k] = $v;
        }
    }
    return $PARAM_MAPPING;
}

?>
