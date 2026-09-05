package main

import (
	"generator/c"
	"generator/fb"
	"os"
	"slices"
	"strings"
)

func main() {
	err := Generate(BindGenOpts{
		Inputs:  []string{"headers/glfw3.h"},
		Defines: []string{"GLFW_INCLUDE_VULKAN"},

		Outputs: []File{
			{Path: "../src/glfw.fb", Module: "glfw"},
		},

		NameMappings: map[string]string{
			"GLFWglproc": "GlProc",
			"GLFWvkproc": "VkProc",

			"GLFWwindowposfun":          "WindowPosFn",
			"GLFWwindowsizefun":         "WindowSizeFn",
			"GLFWwindowclosefun":        "WindowCloseFn",
			"GLFWwindowrefreshfun":      "WindowRefreshFn",
			"GLFWwindowfocusfun":        "WindowFocusFn",
			"GLFWwindowiconifyfun":      "WindowIconifyFn",
			"GLFWwindowmaximizefun":     "WindowMaximizeFn",
			"GLFWframebuffersizefun":    "FramebufferSizeFn",
			"GLFWwindowcontentscalefun": "WindowContentScaleFn",
			"GLFWmousebuttonfun":        "MouseButtonFn",
			"GLFWcursorposfun":          "CursorPosFn",
			"GLFWcursorenterfun":        "CursorEnterFn",
			"GLFWcharmodsfun":           "CharModsFn",

			"Vidmode":      "VidMode",
			"Gammaramp":    "GammaRamp",
			"Gamepadstate": "GamepadState",
		},

		MacroEnums: []MacroEnum{
			{Name: "Hat", Bitfield: true, Prefix: "GLFW_HAT_"},
			{Name: "Key", Type: "i32", Prefix: "GLFW_KEY_"},
			{Name: "Mod", Bitfield: true, Prefix: "GLFW_MOD_"},
			{Name: "MouseButton", Prefix: "GLFW_MOUSE_BUTTON_"},
			{Name: "Joystick", Prefix: "GLFW_JOYSTICK_"},
			{Name: "GamepadButton", Prefix: "GLFW_GAMEPAD_BUTTON_"},
			{Name: "GamepadAxis", Prefix: "GLFW_GAMEPAD_AXIS_"},
			{Name: "Api", Prefix: "GLFW_", Suffix: "_API"},
			{Name: "GlProfile", Prefix: "GLFW_OPENGL_", Suffix: "_PROFILE"},
			{Name: "CursorState", Prefix: "GLFW_CURSOR_"},
			{Name: "AnglePlatform", Prefix: "GLFW_ANGLE_PLATFORM_TYPE_"},
			{Name: "CursorShape", Prefix: "GLFW_", Suffix: "_CURSOR"},

			{Name: "Bool", Type: "i32", Exact: []string{"GLFW_TRUE", "GLFW_FALSE"}},
			{Name: "Action", Type: "i32", Exact: []string{"GLFW_PRESS", "GLFW_RELEASE", "GLFW_REPEAT"}},
			{Name: "DeviceEvent", Type: "i32", Exact: []string{"GLFW_CONNECTED", "GLFW_DISCONNECTED"}},
			{Name: "InputMode", Type: "i32", Exact: []string{"GLFW_CURSOR", "GLFW_STICKY_KEYS", "GLFW_STICKY_MOUSE_BUTTONS", "GLFW_LOCK_KEY_MODS", "GLFW_RAW_MOUSE_MOTION"}},

			{Name: "InitHint", Type: "i32", Exact: []string{
				"GLFW_JOYSTICK_HAT_BUTTONS",
				"GLFW_ANGLE_PLATFORM_TYPE",
				"GLFW_PLATFORM",
				"GLFW_COCOA_CHDIR_RESOURCES",
				"GLFW_COCOA_MENUBAR",
				"GLFW_X11_XCB_VULKAN_SURFACE",
				"GLFW_WAYLAND_LIBDECOR",
			}},

			{Name: "WindowHint", Type: "i32", Exact: []string{
				"GLFW_FOCUSED",
				"GLFW_ICONIFIED",
				"GLFW_RESIZABLE",
				"GLFW_VISIBLE",
				"GLFW_DECORATED",
				"GLFW_AUTO_ICONIFY",
				"GLFW_FLOATING",
				"GLFW_MAXIMIZED",
				"GLFW_CENTER_CURSOR",
				"GLFW_TRANSPARENT_FRAMEBUFFER",
				"GLFW_HOVERED",
				"GLFW_FOCUS_ON_SHOW",
				"GLFW_MOUSE_PASSTHROUGH",
				"GLFW_POSITION_X",
				"GLFW_POSITION_Y",
				"GLFW_RED_BITS",
				"GLFW_GREEN_BITS",
				"GLFW_BLUE_BITS",
				"GLFW_ALPHA_BITS",
				"GLFW_DEPTH_BITS",
				"GLFW_STENCIL_BITS",
				"GLFW_ACCUM_RED_BITS",
				"GLFW_ACCUM_GREEN_BITS",
				"GLFW_ACCUM_BLUE_BITS",
				"GLFW_ACCUM_ALPHA_BITS",
				"GLFW_AUX_BUFFERS",
				"GLFW_STEREO",
				"GLFW_SAMPLES",
				"GLFW_SRGB_CAPABLE",
				"GLFW_REFRESH_RATE",
				"GLFW_DOUBLEBUFFER",
				"GLFW_CLIENT_API",
				"GLFW_CONTEXT_VERSION_MAJOR",
				"GLFW_CONTEXT_VERSION_MINOR",
				"GLFW_CONTEXT_REVISION",
				"GLFW_CONTEXT_ROBUSTNESS",
				"GLFW_OPENGL_FORWARD_COMPAT",
				"GLFW_CONTEXT_DEBUG",
				"GLFW_OPENGL_DEBUG_CONTEXT",
				"GLFW_OPENGL_PROFILE",
				"GLFW_CONTEXT_RELEASE_BEHAVIOR",
				"GLFW_CONTEXT_NO_ERROR",
				"GLFW_CONTEXT_CREATION_API",
				"GLFW_SCALE_TO_MONITOR",
				"GLFW_SCALE_FRAMEBUFFER",
				"GLFW_COCOA_RETINA_FRAMEBUFFER",
				"GLFW_COCOA_FRAME_NAME",
				"GLFW_COCOA_GRAPHICS_SWITCHING",
				"GLFW_X11_CLASS_NAME",
				"GLFW_X11_INSTANCE_NAME",
				"GLFW_WIN32_KEYBOARD_MENU",
				"GLFW_WIN32_SHOWDEFAULT",
				"GLFW_WAYLAND_APP_ID",
			}},

			{Name: "ErrorCode", Type: "i32", Exact: []string{
				"GLFW_NO_ERROR",
				"GLFW_NOT_INITIALIZED",
				"GLFW_NO_CURRENT_CONTEXT",
				"GLFW_INVALID_ENUM",
				"GLFW_INVALID_VALUE",
				"GLFW_OUT_OF_MEMORY",
				"GLFW_API_UNAVAILABLE",
				"GLFW_VERSION_UNAVAILABLE",
				"GLFW_PLATFORM_ERROR",
				"GLFW_FORMAT_UNAVAILABLE",
				"GLFW_NO_WINDOW_CONTEXT",
				"GLFW_CURSOR_UNAVAILABLE",
				"GLFW_FEATURE_UNAVAILABLE",
				"GLFW_FEATURE_UNIMPLEMENTED",
				"GLFW_PLATFORM_UNAVAILABLE",
			}},
		},

		FilterAlias: func(node *c.Node) bool {
			return strings.HasPrefix(node.Name, "GLFW")
		},
		FilterStruct: func(node *c.Node) bool {
			return strings.HasPrefix(node.Name, "GLFW")
		},
		FilterFunc: func(node *c.Node) bool {
			return strings.HasPrefix(node.Name, "glfw")
		},

		ParseType: func(str string) fb.Type {
			switch str {
			case "VkInstance", "VkPhysicalDevice", "VkSurfaceKHR":
				return &fb.PointerType{Mutable: true, Pointee: &fb.SimpleType{Text: "void"}}

			case "VkAllocationCallbacks":
				return &fb.PointerType{Pointee: &fb.SimpleType{Text: "void"}}

			case "VkResult":
				return &fb.SimpleType{Text: "i32"}

			case "PFN_vkGetInstanceProcAddr":
				return &fb.FuncType{
					Params: []fb.Param{
						{Type: &fb.PointerType{Mutable: true, Pointee: &fb.SimpleType{Text: "void"}}},
						{Type: &fb.PointerType{Pointee: &fb.SimpleType{Text: "u8"}}},
					},
					Returns: &fb.PointerType{Pointee: &fb.SimpleType{Text: "void"}},
				}

			default:
				return nil
			}
		},

		TransformAlias: func(a *fb.Alias) {
			if before, ok := strings.CutPrefix(a.Name, "GLFW"); ok {
				a.Name = strings.ToUpper(before[:1]) + before[1:]
			}

			if before, ok := strings.CutSuffix(a.Name, "fun"); ok {
				a.Name = before + "Fn"
			}

			switch a.Name {
			case "KeyFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[1].Type = &fb.SimpleType{Text: "Key"}
				typ.Params[3].Type = &fb.SimpleType{Text: "Action"}
				typ.Params[4].Type = &fb.SimpleType{Text: "Mod"}

			case "CharModsFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[2].Type = &fb.SimpleType{Text: "Mod"}

			case "MouseButtonFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[1].Type = &fb.SimpleType{Text: "MouseButton"}
				typ.Params[2].Type = &fb.SimpleType{Text: "Action"}
				typ.Params[3].Type = &fb.SimpleType{Text: "Mod"}

			case "JoystickFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[0].Type = &fb.SimpleType{Text: "Joystick"}
				typ.Params[1].Type = &fb.SimpleType{Text: "DeviceEvent"}

			case "MonitorFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[1].Type = &fb.SimpleType{Text: "DeviceEvent"}

			case "WindowFocusFn", "WindowIconifyFn", "WindowMaximizeFn", "CursorEnterFn":
				typ := a.Type.(*fb.FuncType)
				typ.Params[1].Type = &fb.SimpleType{Text: "Bool"}
			}
		},

		TransformStruct: func(s *fb.Struct) {
			s.Name = strings.TrimPrefix(s.Name, "GLFW")
			s.Name = strings.ToUpper(s.Name[:1]) + s.Name[1:]
		},

		TransformEnum: func(e *fb.Enum) {
			switch e.Name {
			case "GlProfile":
				e.Cases = slices.DeleteFunc(e.Cases, func(cas *fb.Case) bool {
					return cas.Name == "Profile"
				})

			case "Joystick":
				e.Cases = slices.DeleteFunc(e.Cases, func(cas *fb.Case) bool {
					ch := cas.Name[len(cas.Name)-1]
					return ch < '0' || ch > '9'
				})

			case "AnglePlatform":
				if e.Case("D3D11").Value == e.Case("Metal").Value {
					e.Case("D3D11").Name = "MetalOrD3D11"

					e.Cases = slices.DeleteFunc(e.Cases, func(cas *fb.Case) bool {
						return cas.Name == "Metal"
					})
				}
			}

			for _, cas := range e.Cases {
				if cas.Name[0] >= '0' && cas.Name[0] <= '9' {
					cas.Name = "Num" + cas.Name
				} else {
					cas.Name = strings.TrimPrefix(cas.Name, "Glfw")
				}
			}
		},

		TransformFunc: func(f *fb.Func) {
			f.Name = strings.TrimPrefix(f.Name, "glfw_")

			for _, param := range f.Params {
				switch param.Name {
				case "key":
					if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
						typ.Text = "Key"
					}

				case "button":
					if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
						typ.Text = "MouseButton"
					}

				case "jid":
					if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
						typ.Text = "Joystick"
					}

				case "shape":
					if f.Name == "create_standard_cursor" {
						if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
							typ.Text = "CursorShape"
						}
					}

				case "mode":
					if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
						typ.Text = "InputMode"
					}

				case "value":
					if f.Name == "set_window_should_close" {
						if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
							typ.Text = "Bool"
						}
					}

				case "hint":
					if typ, ok := param.Type.(*fb.SimpleType); ok && typ.Text == "i32" {
						if strings.Contains(f.Name, "init") {
							typ.Text = "InitHint"
						} else {
							typ.Text = "WindowHint"
						}
					}
				}
			}

			switch f.Name {
			case "get_error":
				f.Returns = &fb.SimpleType{Text: "ErrorCode"}

			case "joystick_present", "vulkan_supported", "joystick_is_gamepad",
				"window_should_close", "extension_supported", "platform_supported",
				"raw_mouse_motion_supported", "get_physical_device_presentation_support":
				f.Returns = &fb.SimpleType{Text: "Bool"}
			}

			if len(f.Params) > 0 {
				if p, ok := f.Params[0].Type.(*fb.PointerType); ok {
					if d, ok := p.Pointee.(*fb.DeclType); ok {
						if s, ok := d.Decl.(*fb.Struct); ok {
							if strings.Contains(f.Name, strings.ToLower(s.Name)) {
								f.MethodName = strings.ReplaceAll(f.Name, strings.ToLower(s.Name), "")
								f.MethodName = strings.ReplaceAll(f.MethodName, "__", "_")
								f.MethodName = strings.Trim(f.MethodName, "_")

								f.ReceiverIndex = 0
							}
						}
					}
				}
			}
		},
	})

	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
