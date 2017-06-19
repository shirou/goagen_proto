goagen proto
====================

This package provides a `goa <https://goa.design/>`_ generator for gRPC.

Status
------------------

Very alpha. you can not use in the production.

- Enum is not work
- Required is not work
- Nested type is not work
- Message field order is alphabetical. Should set explicit "Order"?

How to generate proto from your design
---------------------------------------------

At first, you have to do *go get*

::

  % go get github.com/shirou/goagen_proto

Then, you can **goagen gen** with your design.

::

  % goagen gen --pkg-path=github.com/shirou/goagen_proto -d github.com/some/your/great/design

**api.proto** file will be generated.


Type Definition
~~~~~~~~~~~~~~~~~

protobuf has a valious kind of types. But goa can not specify the type.

You can set **Metadata("struct:field:grpctype")** to specify type like this.

::

		Attribute("age", Integer, "age", func() {
			Metadata("struct:field:grpctype", "uint32")
		})


Example
---------------------

See example directory.



LICENSE
---------------------

MIT License
