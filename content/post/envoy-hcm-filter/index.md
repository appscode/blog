---
title: HTTP Filter Creation in Envoy
date: "2024-12-13"
weight: 30
authors:
- Tauhedul Islam
tags:
- envoy
- http-filters
- http
- proxy
---

## Envoy HTTP Filter Creation
In Envoy, we may use HTTP Connection Manager(HCM) for the HTTP based communication. In this filter, we can add some extra sub-filters, which
are called HTTP filters.\
In this blog, we'll write about how to create a new HCM filter.\
\
As there's no official documentation about this, we can't explain the design philosophy, we'll only mention what changes 
should be done and what should be added to implement a new HCM filter and make it workable.

- ### Proto
  At first, we need to build the proto file for the new HCM filter. For this, create a new directory like `envoy/api/envoy/extensions/filters/http/new_hcm_filter/v3`
  Here replace the `new_hcm_filter` with your filter name.\
  Inside this directory, create a `BUILD` file and a `new_hcm_filter.proto` file. In the proto file we need to specify the
  package name which should be something like `envoy.extensions.filters.http.new_hcm_filter.v3`, import the necessary files
  and define necessary options for `go_package`, `java_package`, etc., and create a message something like `NewHCMFilter`.
  We may follow any other http filter's proto file as a reference, for example we may follow the `router` filter.\
  Inside the `NewHCMFilter` message we may add any required parameter we want to set from the Envoy configuration yaml.
  Update the build file accordingly to support the imported file, follow other hcm filters for any help. Most of them has
  similar kind of building dependency and structure.
- ### Filter Implementation
  Now, we need to implement the filter class and its config class maintaining the HCM filter protocol.\
  Each HCM filter should inherit the `PassThroughFilter` class from the `source/extensions/filters/http/common/pass_through_filter.h` header file, 
  if it's meant to be used as both read and write filter.\
  In this filter we need to implement different functions from the `PassThroughFilter` class to have the functionalities of HCM
  filter. Some of them are `decodeHeaders`, `decodeBody`, `encodeHeaders`, `encodeBody`, etc. We'll get the header of a request packet
  going from the downstream to upstream in the `decodeHeaders` function, and the header of the response packet from the upstream to downstream
  from the `encodeHeaders` function. Same for `decodeBody` and `encodeBody` functions.\
  \
  In the config file, we need to implement a FilterConfigFactory class that will be responsible to register the new filter in the filter chain.
  Let's name this class as `NewHCMFilterConfigFactory` and inherit the `FactoryBase` class with specifying the template class name like
  `FactoryBas<envoy::extensions::filters::http::elastic_search::v3::NewHCMFilter>`.\
  Inside this class, implement the Constructor and override the `createFilterFactoryFromProtoTyped` function to create and register the filter class.
  You may also override the `isTerminalFilterByProtoTyped` function and return `true` if the filter is a terminal filter. By default, it is false if you
  don't implement. A terminal filter is a filter which must be at the end of the filter chain. Generally, a filter shouldn't be a terminal filter.\
  For detailed understanding you may follow other HCM Filter's code as reference. Specially for the config part.
- ### Adding References
  To make it completely workable, we need to add the reference of the new filter in some places. Here are the files where we need to add the new filter:
  1. Add the proto build package in the `envoy/api/BUILD` file. Inside this file add the `//envoy/extensions/filters/http/new_hcm_filter/v3:pkg` in the `v3_protos`.
  2. In the `envoy/source/extensions/extensions_build_config.bzl` file's `Extensions` dictionary, add teh new hcm filter maintaining the format `"envoy.filters.http.elastic_search": "//source/extensions/filters/http/elastic_search_filter:elastic_search_config"`.
  3. in the `source/extensions/extensions_metadata.yaml` file, insert the metadata for the new hcm filter maintaining the format:\
    ```
    envoy.filters.http.new_hcm_filter:
     categories:
       - envoy.filters.http
     security_posture: robust_to_untrusted_downstream
     status: stable
     type_urls:
       - envoy.extensions.filters.http.new_hcm_filter.v3.NewHCMFilter
    ```
  4. In the `source/extensions/filters/http/well_known_names.h` file, add the name for the new hcm filter inside the `HttpFilterNameValues` class as a const string.\
    Example Format: `const std::string NewHCMFilter = "envoy.filters.http.new_hcm_filter"`.


### Example config yaml
```yaml
http_filters:
  - name: envoy.filters.http.new_hcm_filter
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.new_hcm_filter.v3.NewHCMFilter
  - name: envoy.filters.http.router
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```


### Conclusion
As Envoy doesn't have any official documentation about creating a new HCM filter till now, the best way is to follow the other implemented 
HCM filters as reference and maintaining their naming and structural formats.\
In this blog we tried to sum up the mandatory changes that needs to be done to implement a new HCM filter and add it with the filter chain. This might change with time 
if the envoy team changes their basic structure. But you might still get help from this.
    
       